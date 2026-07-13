package http

import (
	"github.com/acmhot100/server/internal/model"
	"github.com/acmhot100/server/internal/repository"
)

type tagDTO struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type problemSummaryDTO struct {
	ID            string   `json:"id"`
	Slug          string   `json:"slug"`
	OrderIndex    int      `json:"order_index"`
	Title         string   `json:"title"`
	Difficulty    string   `json:"difficulty"`
	Tags          []tagDTO `json:"tags"`
	ProgressState *string  `json:"progress_state"`
}

type problemListDTO struct {
	Items    []problemSummaryDTO `json:"items"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

type sampleCaseDTO struct {
	ID             string `json:"id"`
	OrderIndex     int    `json:"order_index"`
	InputData      string `json:"input_data"`
	ExpectedOutput string `json:"expected_output"`
	ExplanationMD  string `json:"explanation_md"`
}

type problemDetailDTO struct {
	ID             string          `json:"id"`
	Slug           string          `json:"slug"`
	OrderIndex     int             `json:"order_index"`
	Title          string          `json:"title"`
	Difficulty     string          `json:"difficulty"`
	Stage          string          `json:"stage"`
	Tags           []tagDTO        `json:"tags"`
	ProgressState  *string         `json:"progress_state"`
	StatementMD    string          `json:"statement_md"`
	InputFormatMD  string          `json:"input_format_md"`
	OutputFormatMD string          `json:"output_format_md"`
	ConstraintsMD  string          `json:"constraints_md"`
	HintsMD        string          `json:"hints_md"`
	TimeLimitMs    int             `json:"time_limit_ms"`
	MemoryLimitKb  int             `json:"memory_limit_kb"`
	SampleCases    []sampleCaseDTO `json:"sample_cases"`
}

type languageDTO struct {
	Key            string `json:"key"`
	DisplayName    string `json:"display_name"`
	EditorLanguage string `json:"editor_language"`
	SourceTemplate string `json:"source_template"`
}

type navigationDTO struct {
	Prev *navigationItemDTO `json:"prev"`
	Next *navigationItemDTO `json:"next"`
}

type navigationItemDTO struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

func newProblemSummaryDTO(problem repository.ProblemWithStatus, authenticated bool) problemSummaryDTO {
	var progressState *string
	if authenticated {
		state := problem.State
		if state == "" {
			state = model.ProgressNotStarted
		}
		progressState = &state
	}
	return problemSummaryDTO{
		ID:            problem.ID,
		Slug:          problem.Slug,
		OrderIndex:    problem.OrderIndex,
		Title:         problem.Title,
		Difficulty:    problem.Difficulty,
		Tags:          newTagDTOs(problem.Tags),
		ProgressState: progressState,
	}
}

func newProblemDetailDTO(problem model.Problem, samples []model.TestCase, progressState *string) problemDetailDTO {
	sampleDTOs := make([]sampleCaseDTO, len(samples))
	for i, sample := range samples {
		sampleDTOs[i] = sampleCaseDTO{
			ID:             sample.ID,
			OrderIndex:     sample.OrderIndex,
			InputData:      sample.InputData,
			ExpectedOutput: sample.ExpectedOutput,
			ExplanationMD:  sample.ExplanationMD,
		}
	}
	return problemDetailDTO{
		ID:             problem.ID,
		Slug:           problem.Slug,
		OrderIndex:     problem.OrderIndex,
		Title:          problem.Title,
		Difficulty:     problem.Difficulty,
		Stage:          problem.Stage,
		Tags:           newTagDTOs(problem.Tags),
		ProgressState:  progressState,
		StatementMD:    problem.StatementMD,
		InputFormatMD:  problem.InputFormatMD,
		OutputFormatMD: problem.OutputFormatMD,
		ConstraintsMD:  problem.ConstraintsMD,
		HintsMD:        problem.HintsMD,
		TimeLimitMs:    problem.TimeLimitMs,
		MemoryLimitKb:  problem.MemoryLimitKb,
		SampleCases:    sampleDTOs,
	}
}

func newLanguageDTO(language model.LanguageConfig) languageDTO {
	return languageDTO{
		Key:            language.Key,
		DisplayName:    language.DisplayName,
		EditorLanguage: language.EditorLanguage,
		SourceTemplate: language.SourceTemplate,
	}
}

func newTagDTOs(tags []model.Tag) []tagDTO {
	dtos := make([]tagDTO, len(tags))
	for i, tag := range tags {
		dtos[i] = tagDTO{Slug: tag.Slug, Name: tag.Name}
	}
	return dtos
}
