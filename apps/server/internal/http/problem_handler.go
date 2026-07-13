package http

import (
	"net/http"
	"strconv"

	"github.com/acmhot100/server/internal/model"
	"github.com/acmhot100/server/internal/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// apiError writes the standard API error response.
func apiError(c *gin.Context, status int, code, message string) {
	requestID, _ := c.Get("request_id")
	rid, _ := requestID.(string)
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
		"request_id": rid,
	})
}

// getUserID extracts the authenticated user ID from the context.
func getUserID(c *gin.Context) string {
	userID, _ := c.Get("userID")
	id, _ := userID.(string)
	return id
}

// listProblems handles GET /api/v1/problems.
func listProblems(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 20
		}
		if pageSize > 100 {
			pageSize = 100
		}

		userID := getUserID(c)
		problems, total, err := repository.ListProblems(
			db,
			c.Query("q"),
			c.Query("difficulty"),
			c.Query("tag"),
			c.Query("state"),
			userID,
			page,
			pageSize,
		)
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list problems")
			return
		}

		items := make([]problemSummaryDTO, len(problems))
		for i, problem := range problems {
			items[i] = newProblemSummaryDTO(problem, userID != "")
		}
		c.JSON(http.StatusOK, problemListDTO{
			Items:    items,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		})
	}
}

// getProblem handles GET /api/v1/problems/:slug.
func getProblem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		problem, err := repository.GetProblemBySlug(db, c.Param("slug"))
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get problem")
			return
		}
		if problem == nil || !problem.IsPublished {
			apiError(c, http.StatusNotFound, "NOT_FOUND", "problem not found")
			return
		}

		samples, err := repository.GetSampleCases(db, problem.ID)
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get sample cases")
			return
		}

		var progressState *string
		if userID := getUserID(c); userID != "" {
			progress, err := repository.GetProgress(db, userID, problem.ID)
			if err != nil {
				apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get progress")
				return
			}
			state := model.ProgressNotStarted
			if progress != nil {
				state = progress.State
			}
			progressState = &state
		}

		c.JSON(http.StatusOK, newProblemDetailDTO(*problem, samples, progressState))
	}
}

// getProblemNavigation handles GET /api/v1/problems/:slug/navigation.
func getProblemNavigation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		prev, next, err := repository.GetProblemNavigation(db, c.Param("slug"))
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get navigation")
			return
		}

		response := navigationDTO{}
		if prev != nil {
			response.Prev = &navigationItemDTO{Slug: prev.Slug, Title: prev.Title}
		}
		if next != nil {
			response.Next = &navigationItemDTO{Slug: next.Slug, Title: next.Title}
		}
		c.JSON(http.StatusOK, response)
	}
}

// listTags handles GET /api/v1/tags.
func listTags(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tags, err := repository.ListTags(db)
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list tags")
			return
		}
		c.JSON(http.StatusOK, newTagDTOs(tags))
	}
}
