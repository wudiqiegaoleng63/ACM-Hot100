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
)

func TestSaveDraftRejectsSourceOver64KiBBeforeDatabaseWork(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock := handlerTestDB(t)
	body, err := json.Marshal(map[string]string{
		"source_code": strings.Repeat("a", maxSourceCodeSize+1),
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "user-1")
		c.Set("request_id", "request-draft")
	})
	router.PUT("/api/v1/problems/:slug/drafts/:language_key", saveDraft(db))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/problems/two-sum-target/drafts/cpp17",
		bytes.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusRequestEntityTooLarge, response.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	apiErrorBody, ok := payload["error"].(map[string]any)
	if !ok || apiErrorBody["code"] != "SOURCE_CODE_TOO_LARGE" {
		t.Fatalf("error = %#v, want SOURCE_CODE_TOO_LARGE", payload["error"])
	}
	if payload["request_id"] != "request-draft" {
		t.Fatalf("request_id = %#v, want request-draft", payload["request_id"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected database work: %v", err)
	}
}

func TestSaveDraftRejectsDisabledLanguage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock := handlerTestDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `language_configs` WHERE `key` = ? AND enabled = ?")).
		WithArgs("disabled", true).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "user-1")
		c.Set("request_id", "request-language")
	})
	router.PUT("/api/v1/problems/:slug/drafts/:language_key", saveDraft(db))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/problems/two-sum-target/drafts/disabled",
		strings.NewReader(`{"source_code":"int main() {}"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusBadRequest, response.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	apiErrorBody, ok := payload["error"].(map[string]any)
	if !ok || apiErrorBody["code"] != "INVALID_LANGUAGE" {
		t.Fatalf("error = %#v, want INVALID_LANGUAGE", payload["error"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
