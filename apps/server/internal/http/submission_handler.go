package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/acmhot100/server/internal/model"
	"github.com/acmhot100/server/internal/queue"
	"github.com/acmhot100/server/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const maxSubmissionRequestSize = maxSourceCodeSize*6 + 2048

type createSubmissionRequest struct {
	LanguageKey string  `json:"language_key" binding:"required"`
	SourceCode  *string `json:"source_code" binding:"required"`
}

type submissionSummaryResponse struct {
	ID          string    `json:"id"`
	ProblemSlug string    `json:"problem_slug"`
	LanguageKey string    `json:"language_key"`
	Status      string    `json:"status"`
	PassedCases int       `json:"passed_cases"`
	TotalCases  int       `json:"total_cases"`
	TimeMs      *int      `json:"time_ms"`
	MemoryKb    *int      `json:"memory_kb"`
	CreatedAt   time.Time `json:"created_at"`
}

type submissionDetailResponse struct {
	ID              string                       `json:"id"`
	ProblemSlug     string                       `json:"problem_slug"`
	LanguageKey     string                       `json:"language_key"`
	SourceCode      string                       `json:"source_code"`
	Status          string                       `json:"status"`
	PassedCases     int                          `json:"passed_cases"`
	TotalCases      int                          `json:"total_cases"`
	TimeMs          *int                         `json:"time_ms"`
	MemoryKb        *int                         `json:"memory_kb"`
	CompilerOutput  string                       `json:"compiler_output"`
	ErrorMessage    string                       `json:"error_message"`
	CaseResults     []caseResultResponse         `json:"case_results"`
	CreatedAt       time.Time                    `json:"created_at"`
	JudgedAt        *time.Time                   `json:"judged_at"`
}

type caseResultResponse struct {
	CaseIndex    int    `json:"case_index"`
	Status       string `json:"status"`
	TimeMs       *int   `json:"time_ms"`
	MemoryKb     *int   `json:"memory_kb"`
	ActualOutput string `json:"actual_output,omitempty"`
	IsSample     bool   `json:"is_sample"`
}

type submissionListResponse struct {
	Items    []submissionSummaryResponse `json:"items"`
	Total    int                         `json:"total"`
	Page     int                         `json:"page"`
	PageSize int                         `json:"page_size"`
}

// createSubmission handles POST /api/v1/problems/:slug/submissions
func createSubmission(db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserID(c)
		if userID == "" {
			apiError(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
			return
		}

		var req createSubmissionRequest
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSubmissionRequestSize)
		if err := c.ShouldBindJSON(&req); err != nil {
			var maxBytesError *http.MaxBytesError
			if errors.As(err, &maxBytesError) {
				apiError(c, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "submission request is too large")
				return
			}
			apiError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
			return
		}
		if req.SourceCode == nil || len(*req.SourceCode) > maxSourceCodeSize {
			apiError(c, http.StatusRequestEntityTooLarge, "SOURCE_CODE_TOO_LARGE", "source code exceeds 64KB limit")
			return
		}

		problem, err := repository.GetProblemBySlug(db, c.Param("slug"))
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to resolve problem")
			return
		}
		if problem == nil || !problem.IsPublished {
			apiError(c, http.StatusNotFound, "NOT_FOUND", "problem not found")
			return
		}

		language, err := repository.GetEnabledLanguageByKey(db, req.LanguageKey)
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to validate language")
			return
		}
		if language == nil {
			apiError(c, http.StatusBadRequest, "INVALID_LANGUAGE", "language is not enabled")
			return
		}

		submission := &model.Submission{
			ID:          uuid.New().String(),
			UserID:      userID,
			ProblemID:   problem.ID,
			LanguageKey: language.Key,
			SourceCode:  *req.SourceCode,
			Status:      model.SubmissionStatusQueued,
		}
		if err := repository.CreateSubmission(db, submission); err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create submission")
			return
		}

		// Enqueue to Redis Stream — failure does NOT delete the MySQL submission.
		// Reconciliation will pick up un-enqueued submissions later.
		if rdb != nil {
			messageID, err := queue.EnqueueSubmission(c.Request.Context(), rdb, submission.ID)
			if err == nil {
				enqueuedAt := time.Now().UTC()
				_ = repository.MarkSubmissionEnqueued(db, submission.ID, messageID, enqueuedAt)
			}
		}

		c.JSON(http.StatusAccepted, gin.H{
			"id":         submission.ID,
			"status":     submission.Status,
			"created_at": submission.CreatedAt,
		})
	}
}

// listSubmissions handles GET /api/v1/submissions
func listSubmissions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserID(c)
		if userID == "" {
			apiError(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
			return
		}

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

		submissions, total, err := repository.ListSubmissions(
			db,
			userID,
			c.Query("problem"),
			c.Query("status"),
			c.Query("language"),
			page,
			pageSize,
		)
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list submissions")
			return
		}

		// Build slug lookup for problem IDs
		items := make([]submissionSummaryResponse, len(submissions))
		for i, sub := range submissions {
			items[i] = submissionSummaryResponse{
				ID:          sub.ID,
				LanguageKey: sub.LanguageKey,
				Status:      sub.Status,
				PassedCases: sub.PassedCases,
				TotalCases:  sub.TotalCases,
				TimeMs:      sub.TimeMs,
				MemoryKb:    sub.MemoryKb,
				CreatedAt:   sub.CreatedAt,
			}
		}

		// Resolve problem slugs for the response
		if len(submissions) > 0 {
			problemIDs := make([]string, 0, len(submissions))
			seen := make(map[string]bool)
			for _, sub := range submissions {
				if !seen[sub.ProblemID] {
					problemIDs = append(problemIDs, sub.ProblemID)
					seen[sub.ProblemID] = true
				}
			}
			var problems []model.Problem
			db.Where("id IN ?", problemIDs).Select("id, slug").Find(&problems)
			slugMap := make(map[string]string, len(problems))
			for _, p := range problems {
				slugMap[p.ID] = p.Slug
			}
			for i, sub := range submissions {
				items[i].ProblemSlug = slugMap[sub.ProblemID]
			}
		}

		c.JSON(http.StatusOK, submissionListResponse{
			Items:    items,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		})
	}
}

// getSubmission handles GET /api/v1/submissions/:id
func getSubmission(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserID(c)
		if userID == "" {
			apiError(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
			return
		}

		submission, err := repository.GetSubmissionForUser(db, c.Param("id"), userID)
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get submission")
			return
		}
		if submission == nil {
			apiError(c, http.StatusNotFound, "NOT_FOUND", "submission not found")
			return
		}

		problemSlug := ""
		if p, err := repository.GetProblemByID(db, submission.ProblemID); err == nil && p != nil {
			problemSlug = p.Slug
		}

		caseResults := make([]caseResultResponse, len(submission.CaseResults))
		for i, cr := range submission.CaseResults {
			resp := caseResultResponse{
				CaseIndex: cr.CaseIndex,
				Status:    cr.Status,
				TimeMs:    cr.TimeMs,
				MemoryKb:  cr.MemoryKb,
				IsSample:  cr.IsSample,
			}
			// Only include actual_output for sample cases; hidden cases must not leak
			if cr.IsSample {
				resp.ActualOutput = cr.ActualOutput
			}
			caseResults[i] = resp
		}

		c.JSON(http.StatusOK, submissionDetailResponse{
			ID:             submission.ID,
			ProblemSlug:    problemSlug,
			LanguageKey:    submission.LanguageKey,
			SourceCode:     submission.SourceCode,
			Status:         submission.Status,
			PassedCases:    submission.PassedCases,
			TotalCases:     submission.TotalCases,
			TimeMs:         submission.TimeMs,
			MemoryKb:       submission.MemoryKb,
			CompilerOutput: submission.CompilerOutput,
			ErrorMessage:   submission.ErrorMessage,
			CaseResults:    caseResults,
			CreatedAt:      submission.CreatedAt,
			JudgedAt:       submission.JudgedAt,
		})
	}
}
