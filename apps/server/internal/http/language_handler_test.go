package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestListLanguagesReturnsOnlyPublicEnabledFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock := handlerTestDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `language_configs` WHERE enabled = ? ORDER BY `key` ASC")).
		WithArgs(true).
		WillReturnRows(sqlmock.NewRows([]string{
			"key", "display_name", "judge0_language_name", "judge0_language_id", "editor_language", "source_template", "enabled",
		}).AddRow(
			"cpp17", "C++17", "C++ (gcc 12.2.0)", 54, "cpp", "int main() {}", true,
		))

	router := gin.New()
	router.GET("/api/v1/languages", listLanguages(db))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/languages", nil)
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	var payload []map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("languages = %d, want 1", len(payload))
	}
	for _, field := range []string{"key", "display_name", "editor_language", "source_template"} {
		if _, ok := payload[0][field]; !ok {
			t.Errorf("response missing %q", field)
		}
	}
	for _, forbidden := range []string{"judge0_language_name", "judge0_language_id", "enabled"} {
		if _, ok := payload[0][forbidden]; ok {
			t.Errorf("response exposes internal field %q", forbidden)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}

func handlerTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open GORM with SQL mock: %v", err)
	}
	return db, mock
}
