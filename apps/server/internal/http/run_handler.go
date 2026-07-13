package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/acmhot100/server/internal/model"
	"github.com/acmhot100/server/internal/queue"
	"github.com/acmhot100/server/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	maxCustomInputSize = 16 * 1024
	runRetention       = 24 * time.Hour
	maxRunRequestSize  = maxSourceCodeSize*6 + maxCustomInputSize*6 + 2048
)

type createSampleRunRequest struct {
	LanguageKey  string  `json:"language_key" binding:"required"`
	SourceCode   *string `json:"source_code" binding:"required"`
	SampleCaseID *string `json:"sample_case_id"`
	CustomInput  *string `json:"custom_input"`
}

type sampleRunResponse struct {
	ID           string     `json:"id"`
	LanguageKey  string     `json:"language_key"`
	SampleCaseID *string    `json:"sample_case_id"`
	InputData    string     `json:"input_data"`
	Status       string     `json:"status"`
	OutputData   string     `json:"output_data"`
	ErrorMessage string     `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	StartedAt    *time.Time `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	ExpiresAt    time.Time  `json:"expires_at"`
}

func createSampleRun(db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserID(c)
		if userID == "" {
			apiError(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
			return
		}

		var req createSampleRunRequest
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRunRequestSize)
		if err := c.ShouldBindJSON(&req); err != nil {
			var maxBytesError *http.MaxBytesError
			if errors.As(err, &maxBytesError) {
				apiError(c, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "run request is too large")
				return
			}
			apiError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
			return
		}
		if req.SourceCode == nil || len(*req.SourceCode) > maxSourceCodeSize {
			apiError(c, http.StatusRequestEntityTooLarge, "SOURCE_CODE_TOO_LARGE", "source code exceeds 64KB limit")
			return
		}
		if (req.SampleCaseID == nil) == (req.CustomInput == nil) {
			apiError(c, http.StatusBadRequest, "INVALID_INPUT_SELECTOR", "provide exactly one sample_case_id or custom_input")
			return
		}
		if req.CustomInput != nil && len(*req.CustomInput) > maxCustomInputSize {
			apiError(c, http.StatusRequestEntityTooLarge, "CUSTOM_INPUT_TOO_LARGE", "custom input exceeds 16KB limit")
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

		inputData := ""
		if req.CustomInput != nil {
			inputData = *req.CustomInput
		} else {
			testCase, err := repository.GetSampleCaseForProblem(db, problem.ID, *req.SampleCaseID)
			if err != nil {
				apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to resolve sample case")
				return
			}
			if testCase == nil {
				apiError(c, http.StatusBadRequest, "INVALID_SAMPLE_CASE", "sample case does not belong to this problem")
				return
			}
			inputData = testCase.InputData
		}

		now := time.Now().UTC()
		run := &model.SampleRun{
			ID:           uuid.New().String(),
			UserID:       userID,
			ProblemID:    problem.ID,
			LanguageKey:  language.Key,
			SampleCaseID: req.SampleCaseID,
			SourceCode:   *req.SourceCode,
			InputData:    inputData,
			Status:       model.SampleRunStatusQueued,
			ExpiresAt:    now.Add(runRetention),
		}
		if err := repository.CreateSampleRun(db, run); err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create sample run")
			return
		}

		messageID, err := queue.EnqueueRun(c.Request.Context(), rdb, run.ID)
		if err != nil {
			_ = repository.MarkSampleRunSystemError(db, run.ID, "failed to enqueue sample run", time.Now().UTC())
			apiError(c, http.StatusServiceUnavailable, "QUEUE_UNAVAILABLE", "failed to enqueue sample run")
			return
		}
		enqueuedAt := time.Now().UTC()
		if err := repository.MarkSampleRunEnqueued(db, run.ID, messageID, enqueuedAt); err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to record run enqueue")
			return
		}
		run.EnqueuedAt = &enqueuedAt

		c.JSON(http.StatusAccepted, newSampleRunResponse(run))
	}
}

func getSampleRun(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserID(c)
		if userID == "" {
			apiError(c, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
			return
		}
		run, err := repository.GetSampleRunForUser(db, c.Param("id"), userID)
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get sample run")
			return
		}
		if run == nil {
			apiError(c, http.StatusNotFound, "NOT_FOUND", "sample run not found")
			return
		}
		c.JSON(http.StatusOK, newSampleRunResponse(run))
	}
}

func newSampleRunResponse(run *model.SampleRun) sampleRunResponse {
	return sampleRunResponse{
		ID:           run.ID,
		LanguageKey:  run.LanguageKey,
		SampleCaseID: run.SampleCaseID,
		InputData:    run.InputData,
		Status:       run.Status,
		OutputData:   run.OutputData,
		ErrorMessage: run.ErrorMessage,
		CreatedAt:    run.CreatedAt,
		UpdatedAt:    run.UpdatedAt,
		StartedAt:    run.StartedAt,
		FinishedAt:   run.FinishedAt,
		ExpiresAt:    run.ExpiresAt,
	}
}
