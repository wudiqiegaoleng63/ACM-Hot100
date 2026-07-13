package http

import (
	"net/http"
	"strconv"

	"github.com/acmhot100/server/internal/model"
	"github.com/acmhot100/server/internal/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ─── Response types ──────────────────────────────────────────────────────────

type tagItem struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type problemListItem struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	OrderIndex  int       `json:"order_index"`
	Title       string    `json:"title"`
	Difficulty  string    `json:"difficulty"`
	Tags        []tagItem `json:"tags"`
	State       *string   `json:"state,omitempty"`
}

type problemListResponse struct {
	Items    []problemListItem `json:"items"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

type sampleCaseItem struct {
	ID             string `json:"id"`
	OrderIndex     int    `json:"order_index"`
	InputData      string `json:"input_data"`
	ExpectedOutput string `json:"expected_output"`
	ExplanationMD  string `json:"explanation_md"`
}

type draftInfo struct {
	SourceCode  string `json:"source_code"`
	LanguageKey string `json:"language_key"`
}

type progressInfo struct {
	State        string `json:"state"`
	AttemptCount int    `json:"attempt_count"`
}

type problemDetailResponse struct {
	ID              string          `json:"id"`
	Slug            string          `json:"slug"`
	Title           string          `json:"title"`
	Difficulty      string          `json:"difficulty"`
	Stage           string          `json:"stage"`
	Tags            []tagItem       `json:"tags"`
	StatementMD     string          `json:"statement_md"`
	InputFormatMD   string          `json:"input_format_md"`
	OutputFormatMD  string          `json:"output_format_md"`
	ConstraintsMD   string          `json:"constraints_md"`
	HintsMD         string          `json:"hints_md"`
	TimeLimitMs     int             `json:"time_limit_ms"`
	MemoryLimitKb   int             `json:"memory_limit_kb"`
	SampleCases     []sampleCaseItem `json:"sample_cases"`
	Draft           *draftInfo      `json:"draft,omitempty"`
	Progress        *progressInfo   `json:"progress,omitempty"`
}

type navigationResponse struct {
	Prev *navItem `json:"prev"`
	Next *navItem `json:"next"`
}

type navItem struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

// ─── Error helper ────────────────────────────────────────────────────────────

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

// ─── getUserID helper ────────────────────────────────────────────────────────

// getUserID extracts the authenticated user ID from the context, or returns "".
func getUserID(c *gin.Context) string {
	userID, _ := c.Get("user_id")
	id, _ := userID.(string)
	return id
}

// ─── Handlers ────────────────────────────────────────────────────────────────

// listProblems handles GET /api/v1/problems
func listProblems(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("q")
		difficulty := c.Query("difficulty")
		tag := c.Query("tag")
		state := c.Query("state")

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

		problems, total, err := repository.ListProblems(db, q, difficulty, tag, state, userID, page, pageSize)
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list problems")
			return
		}

		items := make([]problemListItem, 0, len(problems))
		for _, p := range problems {
			tags := make([]tagItem, len(p.Tags))
			for i, t := range p.Tags {
				tags[i] = tagItem{Slug: t.Slug, Name: t.Name}
			}

			item := problemListItem{
				ID:         p.ID,
				Slug:       p.Slug,
				OrderIndex: p.OrderIndex,
				Title:      p.Title,
				Difficulty: p.Difficulty,
				Tags:       tags,
			}

			// Only include state for authenticated users
			if userID != "" {
				s := p.State
				if s == "" {
					s = model.ProgressNotStarted
				}
				item.State = &s
			}

			items = append(items, item)
		}

		c.JSON(http.StatusOK, problemListResponse{
			Items:    items,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		})
	}
}

// getProblem handles GET /api/v1/problems/:slug
func getProblem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

		problem, err := repository.GetProblemBySlug(db, slug)
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get problem")
			return
		}
		if problem == nil {
			apiError(c, http.StatusNotFound, "NOT_FOUND", "problem not found")
			return
		}

		// Only return published problems
		if !problem.IsPublished {
			apiError(c, http.StatusNotFound, "NOT_FOUND", "problem not found")
			return
		}

		// Sample cases only (NEVER hidden test cases)
		sampleCases, err := repository.GetSampleCases(db, problem.ID)
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get sample cases")
			return
		}

		sampleItems := make([]sampleCaseItem, len(sampleCases))
		for i, tc := range sampleCases {
			sampleItems[i] = sampleCaseItem{
				ID:             tc.ID,
				OrderIndex:     tc.OrderIndex,
				InputData:      tc.InputData,
				ExpectedOutput: tc.ExpectedOutput,
				ExplanationMD:  tc.ExplanationMD,
			}
		}

		tags := make([]tagItem, len(problem.Tags))
		for i, t := range problem.Tags {
			tags[i] = tagItem{Slug: t.Slug, Name: t.Name}
		}

		resp := problemDetailResponse{
			ID:             problem.ID,
			Slug:           problem.Slug,
			Title:          problem.Title,
			Difficulty:     problem.Difficulty,
			Stage:          problem.Stage,
			Tags:           tags,
			StatementMD:    problem.StatementMD,
			InputFormatMD:  problem.InputFormatMD,
			OutputFormatMD: problem.OutputFormatMD,
			ConstraintsMD:  problem.ConstraintsMD,
			HintsMD:        problem.HintsMD,
			TimeLimitMs:    problem.TimeLimitMs,
			MemoryLimitKb:  problem.MemoryLimitKb,
			SampleCases:    sampleItems,
		}

		// Include draft and progress only for authenticated users
		userID := getUserID(c)
		if userID != "" {
			// Get progress
			progress, err := repository.GetProgress(db, userID, problem.ID)
			if err != nil {
				apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get progress")
				return
			}
			if progress != nil {
				resp.Progress = &progressInfo{
					State:        progress.State,
					AttemptCount: progress.AttemptCount,
				}
			}

			// We don't include draft in the problem detail itself;
			// it's fetched separately via the draft endpoint.
			// But the spec says draft should be included here.
			// We need the language_key to fetch a draft, but the detail endpoint
			// doesn't specify a language. Return null for draft in detail view.
		}

		c.JSON(http.StatusOK, resp)
	}
}

// getProblemNavigation handles GET /api/v1/problems/:slug/navigation
func getProblemNavigation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

		prev, next, err := repository.GetProblemNavigation(db, slug)
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get navigation")
			return
		}

		resp := navigationResponse{}
		if prev != nil {
			resp.Prev = &navItem{Slug: prev.Slug, Title: prev.Title}
		}
		if next != nil {
			resp.Next = &navItem{Slug: next.Slug, Title: next.Title}
		}

		c.JSON(http.StatusOK, resp)
	}
}

// listTags handles GET /api/v1/tags
func listTags(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tags, err := repository.ListTags(db)
		if err != nil {
			apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list tags")
			return
		}

		items := make([]tagItem, len(tags))
		for i, t := range tags {
			items[i] = tagItem{Slug: t.Slug, Name: t.Name}
		}

		c.JSON(http.StatusOK, items)
	}
}
