package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func TestCreateSubmissionRejectsOversizedSource(t *testing.T) {
	db, mock := handlerTestDB(t)
	body, err := json.Marshal(map[string]interface{}{
		"language_key": "cpp17",
		"source_code":  strings.Repeat("a", maxSourceCodeSize+1),
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	router := authenticatedSubmissionRouter(db, nil)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/problems/two-sum-target/submissions", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusRequestEntityTooLarge, response.Body.String())
	}
	assertSubmissionErrorCode(t, response, "SOURCE_CODE_TOO_LARGE")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB work: %v", err)
	}
}

func TestCreateSubmissionRejectsMissingSourceCode(t *testing.T) {
	db, mock := handlerTestDB(t)
	body := strings.NewReader(`{"language_key":"cpp17"}`)

	router := authenticatedSubmissionRouter(db, nil)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/problems/two-sum-target/submissions", body)
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusBadRequest, response.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB work: %v", err)
	}
}

func TestGetSubmissionReturnsNotFoundForAnotherUser(t *testing.T) {
	db, mock := handlerTestDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `submissions` WHERE id = ? AND user_id = ? ORDER BY `submissions`.`id` LIMIT ?")).
		WithArgs("sub-other", "user-1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("userID", "user-1") })
	router.GET("/api/v1/submissions/:id", getSubmission(db))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/submissions/sub-other", nil))

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body=%s", response.Code, response.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func TestGetSubmissionRequiresAuth(t *testing.T) {
	db, _ := handlerTestDB(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/submissions/:id", getSubmission(db))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/submissions/sub-1", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func authenticatedSubmissionRouter(db *gorm.DB, rdb *redis.Client) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "user-1")
		c.Set("request_id", "request-submission")
	})
	router.POST("/api/v1/problems/:slug/submissions", createSubmission(db, rdb))
	return router
}

func assertSubmissionErrorCode(t *testing.T, response *httptest.ResponseRecorder, want string) {
	t.Helper()
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if payload.Error.Code != want || payload.RequestID != "request-submission" {
		t.Fatalf("error = %#v, want code=%s request_id=request-submission", payload, want)
	}
}
