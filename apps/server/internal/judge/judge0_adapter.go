package judge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/acmhot100/server/internal/model"
	"github.com/acmhot100/server/internal/repository"
	"gorm.io/gorm"
)

// Judge0Adapter implements the Adapter interface using the Judge0 HTTP API.
type Judge0Adapter struct {
	db      *gorm.DB
	baseURL string
	client  *http.Client
}

// Judge0AdapterConfig holds the configuration for creating a Judge0Adapter.
type Judge0AdapterConfig struct {
	BaseURL        string
	ConnectTimeout time.Duration
	TotalTimeout   time.Duration
}

// NewJudge0Adapter creates a new Judge0 HTTP adapter.
func NewJudge0Adapter(db *gorm.DB, cfg Judge0AdapterConfig) *Judge0Adapter {
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     cfg.ConnectTimeout,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   cfg.TotalTimeout,
	}
	return &Judge0Adapter{
		db:      db,
		baseURL: cfg.BaseURL,
		client:  client,
	}
}

// Judge0 status ID mapping
const (
	judge0StatusAccepted         = 3
	judge0StatusWrongAnswer      = 4
	judge0StatusTimeLimit        = 5
	judge0StatusMemoryLimit      = 6
	judge0StatusRuntimeError     = 7 // SIGSEGV
	judge0StatusCompilationError = 13
	judge0StatusInternalError    = 10
	judge0StatusExecFormatError  = 11
)

// mapJudge0Status maps a Judge0 status ID to our internal status string.
func mapJudge0Status(statusID int) string {
	switch statusID {
	case judge0StatusAccepted:
		return model.SubmissionStatusAccepted
	case judge0StatusWrongAnswer:
		return model.SubmissionStatusWrongAnswer
	case judge0StatusTimeLimit:
		return model.SubmissionStatusTimeLimit
	case judge0StatusMemoryLimit:
		return model.SubmissionStatusMemoryLimit
	case judge0StatusCompilationError:
		return model.SubmissionStatusCompileError
	case judge0StatusRuntimeError, 8, 9: // various runtime errors
		return model.SubmissionStatusRuntimeError
	case judge0StatusInternalError, judge0StatusExecFormatError:
		return model.SubmissionStatusSystemError
	default:
		return model.SubmissionStatusSystemError
	}
}

type judge0SubmissionRequest struct {
	SourceCode     string  `json:"source_code"`
	LanguageID     int     `json:"language_id"`
	Stdin          string  `json:"stdin,omitempty"`
	ExpectedOutput string  `json:"expected_output,omitempty"`
	CPUTimeLimit   float64 `json:"cpu_time_limit,omitempty"`
	MemoryLimit    int     `json:"memory_limit,omitempty"`
	MaxFileSize    int     `json:"max_file_size,omitempty"`
}

type judge0TokenResponse struct {
	Token string `json:"token"`
}

type judge0ResultResponse struct {
	Stdout        *string      `json:"stdout"`
	Stderr        *string      `json:"stderr"`
	CompileOutput *string      `json:"compile_output"`
	Status        judge0Status `json:"status"`
	Time          string       `json:"time"`
	Memory        int          `json:"memory"`
}

type judge0Status struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

// Judge implements the Adapter interface.
// It loads the submission data, submits each test case to Judge0,
// polls for results, compares outputs, and aggregates the final verdict.
func (a *Judge0Adapter) Judge(ctx context.Context, submissionID string) (*JudgeResult, error) {
	// Load submission
	submission, err := repository.GetSubmissionByID(a.db, submissionID)
	if err != nil {
		return nil, fmt.Errorf("load submission: %w", err)
	}
	if submission == nil {
		return FakeSystemErrorResult("submission not found"), nil
	}

	// Load language config to get Judge0 language ID
	language, err := repository.GetEnabledLanguageByKey(a.db, submission.LanguageKey)
	if err != nil {
		return nil, fmt.Errorf("load language: %w", err)
	}
	if language == nil || language.Judge0LanguageID == nil {
		return FakeSystemErrorResult("language has no Judge0 ID"), nil
	}

	// Load problem for time/memory limits
	problem, err := repository.GetProblemByID(a.db, submission.ProblemID)
	if err != nil {
		return nil, fmt.Errorf("load problem: %w", err)
	}
	if problem == nil {
		return FakeSystemErrorResult("problem not found"), nil
	}

	// Load test cases
	testCases, err := repository.GetAllCases(a.db, submission.ProblemID)
	if err != nil {
		return nil, fmt.Errorf("load test cases: %w", err)
	}

	if len(testCases) == 0 {
		return FakeSystemErrorResult("no test cases"), nil
	}

	// Execute each test case
	verdicts := make([]Verdict, len(testCases))
	passedCases := 0
	totalTimeMs := 0
	peakMemoryKb := 0
	firstNonACStatus := ""
	compilerOutput := ""
	errorMessage := ""

	for i, tc := range testCases {
		req := judge0SubmissionRequest{
			SourceCode:     submission.SourceCode,
			LanguageID:     *language.Judge0LanguageID,
			Stdin:          tc.InputData,
			ExpectedOutput: tc.ExpectedOutput,
			CPUTimeLimit:   float64(problem.TimeLimitMs) / 1000.0,
			MemoryLimit:    problem.MemoryLimitKb,
			MaxFileSize:    8 * 1024, // 8KB max output
		}

		result, err := a.submitAndWait(ctx, req)
		if err != nil {
			verdicts[i] = Verdict{Status: model.SubmissionStatusSystemError, ActualOutput: err.Error()}
			if firstNonACStatus == "" {
				firstNonACStatus = model.SubmissionStatusSystemError
			}
			continue
		}

		// Map status
		status := mapJudge0Status(result.Status.ID)

		// Parse time
		timeMs := parseTimeMs(result.Time)

		// Capture sanitized diagnostics without exposing host paths.
		if i == 0 && result.CompileOutput != nil {
			compilerOutput = TruncateOutput(SanitizePath(*result.CompileOutput))
		}
		if errorMessage == "" && result.Stderr != nil {
			errorMessage = TruncateOutput(SanitizePath(*result.Stderr))
		}

		// For CE, stop immediately
		if status == model.SubmissionStatusCompileError {
			return &JudgeResult{
				Status:         model.SubmissionStatusCompileError,
				CompilerOutput: compilerOutput,
				ErrorMessage:   errorMessage,
				CaseResults:    []Verdict{{Status: model.SubmissionStatusCompileError}},
			}, nil
		}

		output := ""
		if result.Stdout != nil {
			output = *result.Stdout
		}

		verdicts[i] = Verdict{
			Status:       status,
			TimeMs:       timeMs,
			MemoryKb:     result.Memory,
			ActualOutput: TruncateOutput(output),
		}

		totalTimeMs += timeMs
		if result.Memory > peakMemoryKb {
			peakMemoryKb = result.Memory
		}

		if status == model.SubmissionStatusAccepted {
			passedCases++
		} else if firstNonACStatus == "" {
			firstNonACStatus = status
		}

		// For TLE/MLE/RE, stop early (remaining cases won't change outcome)
		if status == model.SubmissionStatusTimeLimit ||
			status == model.SubmissionStatusMemoryLimit ||
			status == model.SubmissionStatusRuntimeError {
			// Mark remaining cases as skipped
			for j := i + 1; j < len(testCases); j++ {
				verdicts[j] = Verdict{Status: status}
			}
			break
		}
	}

	// Determine overall status
	overallStatus := model.SubmissionStatusAccepted
	if firstNonACStatus != "" {
		overallStatus = firstNonACStatus
	}

	return &JudgeResult{
		Status:         overallStatus,
		PassedCases:    passedCases,
		TotalCases:     len(testCases),
		TotalTimeMs:    totalTimeMs,
		PeakMemoryKb:   peakMemoryKb,
		CompilerOutput: compilerOutput,
		ErrorMessage:   errorMessage,
		CaseResults:    verdicts,
	}, nil
}

func (a *Judge0Adapter) submitAndWait(ctx context.Context, req judge0SubmissionRequest) (*judge0ResultResponse, error) {
	// Submit
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/submissions?base64_encoded=false", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("submit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("submit status %d: %s", resp.StatusCode, string(respBody))
	}

	var tokenResp judge0TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}

	// Poll for result
	pollInterval := 200 * time.Millisecond
	maxPollTime := 30 * time.Second
	deadline := time.Now().Add(maxPollTime)

	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("poll timeout after %v", maxPollTime)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
		}

		result, done, err := a.pollResult(ctx, tokenResp.Token)
		if err != nil {
			return nil, err
		}
		if done {
			return result, nil
		}
	}
}

func (a *Judge0Adapter) pollResult(ctx context.Context, token string) (*judge0ResultResponse, bool, error) {
	url := fmt.Sprintf("%s/submissions/%s?base64_encoded=false&fields=stdout,stderr,compile_output,status,time,memory", a.baseURL, token)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("poll status %d", resp.StatusCode)
	}

	var result judge0ResultResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, false, fmt.Errorf("decode result: %w", err)
	}

	// Status 1 (In Queue) or 2 (Processing) means not done yet
	if result.Status.ID == 1 || result.Status.ID == 2 {
		return nil, false, nil
	}

	return &result, true, nil
}

func parseTimeMs(timeStr string) int {
	if timeStr == "" {
		return 0
	}
	var seconds float64
	fmt.Sscanf(timeStr, "%f", &seconds)
	return int(seconds * 1000)
}
