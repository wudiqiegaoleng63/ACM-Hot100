package http

import (
	"net/http"

	"github.com/acmhot100/server/internal/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const maxSourceCodeSize = 64 * 1024 // 64KB

type saveDraftRequest struct {
	SourceCode string `json:"source_code" binding:"required"`
}

type draftResponse struct {
	SourceCode  string `json:"source_code"`
	LanguageKey string `json:"language_key"`
	UpdatedAt   string `json:"updated_at"`
}

// saveDraft handles PUT /api/v1/problems/:slug/drafts/:language_key
func saveDraft(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserID(c)
		if userID == "" {
			apiError(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
			return
		}

		slug := c.Param("slug")
		languageKey := c.Param("language_key")

		// Resolve problem slug to ID
		problem, err := repository.GetProblemBySlug(db, slug)
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to resolve problem")
			return
		}
		if problem == nil || !problem.IsPublished {
			apiError(c, http.StatusNotFound, "NOT_FOUND", "problem not found")
			return
		}

		var req saveDraftRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apiError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
			return
		}

		// Validate source code size
		if len(req.SourceCode) > maxSourceCodeSize {
			apiError(c, http.StatusBadRequest, "BAD_REQUEST", "source code exceeds 64KB limit")
			return
		}

		if err := repository.UpsertDraft(db, userID, problem.ID, languageKey, req.SourceCode); err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to save draft")
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "draft saved"})
	}
}

// getDraft handles GET /api/v1/problems/:slug/drafts/:language_key
func getDraft(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserID(c)
		if userID == "" {
			apiError(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
			return
		}

		slug := c.Param("slug")
		languageKey := c.Param("language_key")

		// Resolve problem slug to ID
		problem, err := repository.GetProblemBySlug(db, slug)
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to resolve problem")
			return
		}
		if problem == nil || !problem.IsPublished {
			apiError(c, http.StatusNotFound, "NOT_FOUND", "problem not found")
			return
		}

		draft, err := repository.GetDraft(db, userID, problem.ID, languageKey)
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get draft")
			return
		}
		if draft == nil {
			apiError(c, http.StatusNotFound, "NOT_FOUND", "draft not found")
			return
		}

		c.JSON(http.StatusOK, draftResponse{
			SourceCode:  draft.SourceCode,
			LanguageKey: draft.LanguageKey,
			UpdatedAt:   draft.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}
