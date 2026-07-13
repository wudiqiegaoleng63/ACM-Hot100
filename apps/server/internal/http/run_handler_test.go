package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func TestCreateSampleRunRejectsOversizedSourceBeforeDependencies(t *testing.T) {
	db, mock := handlerTestDB(t)
	body, err := json.Marshal(map[string]interface{}{
		"language_key": "cpp17",
		"source_code":  strings.Repeat("a", maxSourceCodeSize+1),
		"custom_input": "",
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	router := authenticatedRunRouter(db, nil)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/problems/two-sum-target/run", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusRequestEntityTooLarge, response.Body.String())
	}
	assertRunErrorCode(t, response, "SOURCE_CODE_TOO_LARGE")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB work: %v", err)
	}
}

func TestCreateSampleRunRejectsOversizedCustomInputBeforeDependencies(t *testing.T) {
	db, mock := handlerTestDB(t)
	body, err := json.Marshal(map[string]interface{}{
		"language_key": "cpp17",
		"source_code":  "int main() {}",
		"custom_input": strings.Repeat("x", maxCustomInputSize+1),
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	router := authenticatedRunRouter(db, nil)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/problems/two-sum-target/run", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusRequestEntityTooLarge, response.Body.String())
	}
	assertRunErrorCode(t, response, "CUSTOM_INPUT_TOO_LARGE")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB work: %v", err)
	}
}

func TestCreateSampleRunRequiresExactlyOneInputSelector(t *testing.T) {
	db, mock := handlerTestDB(t)
	body := strings.NewReader(`{"language_key":"cpp17","source_code":"int main() {}"}`)

	router := authenticatedRunRouter(db, nil)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/problems/two-sum-target/run", body)
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusBadRequest, response.Body.String())
	}
	assertRunErrorCode(t, response, "INVALID_INPUT_SELECTOR")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB work: %v", err)
	}
}

func authenticatedRunRouter(db *gorm.DB, rdb *redis.Client) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "user-1")
		c.Set("request_id", "request-run")
	})
	router.POST("/api/v1/problems/:slug/run", createSampleRun(db, rdb))
	return router
}

func assertRunErrorCode(t *testing.T, response *httptest.ResponseRecorder, want string) {
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
	if payload.Error.Code != want || payload.RequestID != "request-run" {
		t.Fatalf("error = %#v, want code=%s request_id=request-run", payload, want)
	}
}
