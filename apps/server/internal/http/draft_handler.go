package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/acmhot100/server/internal/model"
	"github.com/acmhot100/server/internal/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	maxSourceCodeSize   = 64 * 1024 // 64KB
	maxDraftRequestSize = maxSourceCodeSize*6 + 1024
)

type saveDraftRequest struct {
	SourceCode *string `json:"source_code" binding:"required"`
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

		var req saveDraftRequest
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxDraftRequestSize)
		if err := c.ShouldBindJSON(&req); err != nil {
			var maxBytesError *http.MaxBytesError
			if errors.As(err, &maxBytesError) {
				apiError(c, http.StatusRequestEntityTooLarge, "SOURCE_CODE_TOO_LARGE", "source code exceeds 64KB limit")
				return
			}
			apiError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
			return
		}
		if req.SourceCode == nil {
			apiError(c, http.StatusBadRequest, "BAD_REQUEST", "source_code is required")
			return
		}
		if len(*req.SourceCode) > maxSourceCodeSize {
			apiError(c, http.StatusRequestEntityTooLarge, "SOURCE_CODE_TOO_LARGE", "source code exceeds 64KB limit")
			return
		}

		var languageCount int64
		if err := db.Model(&model.LanguageConfig{}).
			Where("`key` = ? AND enabled = ?", languageKey, true).
			Count(&languageCount).Error; err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to validate language")
			return
		}
		if languageCount == 0 {
			apiError(c, http.StatusBadRequest, "INVALID_LANGUAGE", "language is not enabled")
			return
		}

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

		if err := repository.UpsertDraft(db, userID, problem.ID, languageKey, *req.SourceCode); err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to save draft")
			return
		}
		draft, err := repository.GetDraft(db, userID, problem.ID, languageKey)
		if err != nil || draft == nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to read saved draft")
			return
		}

		c.JSON(http.StatusOK, newDraftResponse(draft))
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

		c.JSON(http.StatusOK, newDraftResponse(draft))
	}
}

func newDraftResponse(draft *model.Draft) draftResponse {
	return draftResponse{
		SourceCode:  draft.SourceCode,
		LanguageKey: draft.LanguageKey,
		UpdatedAt:   draft.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}
