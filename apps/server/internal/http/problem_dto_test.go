package http

import (
	"encoding/json"
	"testing"

	"github.com/acmhot100/server/internal/model"
	"github.com/acmhot100/server/internal/repository"
)

func TestProblemSummaryDTOUsesCanonicalProgressState(t *testing.T) {
	dto := newProblemSummaryDTO(repository.ProblemWithStatus{
		Problem: model.Problem{
			ID:         "problem-1",
			Slug:       "two-sum-target",
			OrderIndex: 1,
			Title:      "两数目标和",
			Difficulty: model.DifficultyEasy,
		},
		State: model.ProgressAttempted,
	}, true)
	if dto.ProgressState == nil || *dto.ProgressState != model.ProgressAttempted {
		t.Fatalf("progress_state = %v, want %s", dto.ProgressState, model.ProgressAttempted)
	}
}

func TestProblemDetailDTOContract(t *testing.T) {
	progressState := model.ProgressAttempted
	dto := newProblemDetailDTO(
		model.Problem{
			ID:             "problem-1",
			Slug:           "two-sum-target",
			OrderIndex:     1,
			Title:          "两数目标和",
			Difficulty:     model.DifficultyEasy,
			Stage:          model.StageHot100,
			StatementMD:    "题面",
			InputFormatMD:  "输入",
			OutputFormatMD: "输出",
			ConstraintsMD:  "范围",
			HintsMD:        "提示",
			TimeLimitMs:    1000,
			MemoryLimitKb:  262144,
			Tags:           []model.Tag{{Slug: "array", Name: "数组"}},
		},
		[]model.TestCase{{
			ID:             "sample-1",
			OrderIndex:     1,
			InputData:      "4 9\n2 7 11 15\n",
			ExpectedOutput: "1 2\n",
			IsSample:       true,
			ExplanationMD:  "解释",
		}},
		&progressState,
	)

	encoded, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("marshal detail DTO: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatalf("decode detail DTO: %v", err)
	}

	for _, field := range []string{
		"id", "slug", "order_index", "title", "difficulty", "stage", "tags",
		"progress_state", "statement_md", "input_format_md", "output_format_md",
		"constraints_md", "hints_md", "time_limit_ms", "memory_limit_kb", "sample_cases",
	} {
		if _, ok := payload[field]; !ok {
			t.Errorf("detail response missing %q", field)
		}
	}
	for _, forbidden := range []string{"test_cases", "is_published", "memory_limit_mb", "state"} {
		if _, ok := payload[forbidden]; ok {
			t.Errorf("detail response exposes forbidden field %q", forbidden)
		}
	}
	if payload["difficulty"] != model.DifficultyEasy {
		t.Errorf("difficulty = %v, want %s", payload["difficulty"], model.DifficultyEasy)
	}
	if payload["progress_state"] != model.ProgressAttempted {
		t.Errorf("progress_state = %v, want %s", payload["progress_state"], model.ProgressAttempted)
	}

	samples, ok := payload["sample_cases"].([]any)
	if !ok || len(samples) != 1 {
		t.Fatalf("sample_cases = %#v, want one sample", payload["sample_cases"])
	}
	sample := samples[0].(map[string]any)
	if _, ok := sample["is_sample"]; ok {
		t.Error("sample response exposes internal is_sample field")
	}
}

func TestLanguageDTOExcludesJudgeInternals(t *testing.T) {
	dto := newLanguageDTO(model.LanguageConfig{
		Key:                "cpp17",
		DisplayName:        "C++17",
		Judge0LanguageName: "C++ (gcc 12.2.0)",
		Judge0LanguageID:   intPointer(54),
		EditorLanguage:     "cpp",
		SourceTemplate:     "int main() {}",
		Enabled:            true,
	})
	encoded, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("marshal language DTO: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatalf("decode language DTO: %v", err)
	}
	for _, field := range []string{"key", "display_name", "editor_language", "source_template"} {
		if _, ok := payload[field]; !ok {
			t.Errorf("language response missing %q", field)
		}
	}
	for _, forbidden := range []string{"judge0_language_name", "judge0_language_id", "enabled"} {
		if _, ok := payload[forbidden]; ok {
			t.Errorf("language response exposes internal field %q", forbidden)
		}
	}
}

func intPointer(value int) *int { return &value }
