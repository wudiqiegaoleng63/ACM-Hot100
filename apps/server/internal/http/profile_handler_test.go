package http

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestProfileSummaryReturnsAuthenticatedUsersRealProgress(t *testing.T) {
	db, mock := handlerTestDB(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) AS total_problems,")).
		WithArgs("SOLVED", "ATTEMPTED", "NOT_STARTED", "user-1", true).
		WillReturnRows(sqlmock.NewRows([]string{"total_problems", "solved", "attempted", "not_started"}).AddRow(5, 2, 1, 2))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("userID", "user-1") })
	router.GET("/api/v1/profile/summary", getProfileSummary(db))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/profile/summary", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", response.Code, response.Body.String())
	}
	if got := response.Body.String(); got != `{"total_problems":5,"solved":2,"attempted":1,"not_started":2}` {
		t.Fatalf("body = %s", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet SQL expectations: %v", err)
	}
}
