package judge

import "context"

// Verdict represents a single test case result from the judge.
type Verdict struct {
	Status       string // AC, WA, TLE, MLE, RE, CE
	TimeMs       int
	MemoryKb     int
	ActualOutput string // truncated to 8KB
}

// JudgeResult holds the aggregated result for a submission.
type JudgeResult struct {
	Status         string    // AC, WA, TLE, MLE, RE, CE, SYSTEM_ERROR
	PassedCases    int
	TotalCases     int
	TotalTimeMs    int
	PeakMemoryKb  int
	CompilerOutput string   // truncated to 8KB
	CaseResults    []Verdict
}

// Adapter is the small interface for judge execution.
// Workers depend on this interface, not on HTTP details.
type Adapter interface {
	// Judge executes all test cases for a submission and returns the aggregated result.
	Judge(ctx context.Context, submissionID string) (*JudgeResult, error)
}

// IsTerminalSubmissionStatus returns true for statuses that cannot transition further.
func IsTerminalSubmissionStatus(status string) bool {
	switch status {
	case "AC", "WA", "TLE", "MLE", "RE", "CE", "SYSTEM_ERROR":
		return true
	default:
		return false
	}
}
