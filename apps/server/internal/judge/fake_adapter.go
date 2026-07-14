package judge

import "context"

// FakeAdapter is a deterministic mock adapter for testing.
// It always returns the pre-configured verdicts.
type FakeAdapter struct {
	result *JudgeResult
	err    error
}

// NewFakeAdapter creates a mock adapter that returns the given result.
func NewFakeAdapter(result *JudgeResult) *FakeAdapter {
	return &FakeAdapter{result: result}
}

// NewFakeAdapterWithError creates a mock adapter that returns an error.
func NewFakeAdapterWithError(err error) *FakeAdapter {
	return &FakeAdapter{err: err}
}

// Judge implements the Adapter interface.
func (f *FakeAdapter) Judge(_ context.Context, _ string) (*JudgeResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

// --- Pre-built fake results for common verdicts ---

// FakeACResult returns a mock AC result for the given number of test cases.
func FakeACResult(totalCases int) *JudgeResult {
	cases := make([]Verdict, totalCases)
	for i := range cases {
		cases[i] = Verdict{Status: "AC", TimeMs: 10, MemoryKb: 1024}
	}
	return &JudgeResult{
		Status:        "AC",
		PassedCases:   totalCases,
		TotalCases:    totalCases,
		TotalTimeMs:   totalCases * 10,
		PeakMemoryKb: 1024,
		CaseResults:   cases,
	}
}

// FakeWAResult returns a mock WA result (fails at the given case index).
func FakeWAResult(totalCases, failIndex int) *JudgeResult {
	cases := make([]Verdict, totalCases)
	for i := range cases {
		if i < failIndex {
			cases[i] = Verdict{Status: "AC", TimeMs: 10, MemoryKb: 1024}
		} else if i == failIndex {
			cases[i] = Verdict{Status: "WA", TimeMs: 10, MemoryKb: 1024, ActualOutput: "wrong"}
		} else {
			cases[i] = Verdict{Status: "WA", TimeMs: 0, MemoryKb: 0}
		}
	}
	return &JudgeResult{
		Status:       "WA",
		PassedCases:  failIndex,
		TotalCases:   totalCases,
		TotalTimeMs:  failIndex*10 + 10,
		PeakMemoryKb: 1024,
		CaseResults:  cases,
	}
}

// FakeCEResult returns a mock CE result.
func FakeCEResult() *JudgeResult {
	return &JudgeResult{
		Status:         "CE",
		PassedCases:    0,
		TotalCases:     0,
		CompilerOutput: "error: expected ';' at end of input",
	}
}

// FakeTLEResult returns a mock TLE result (fails at the given case index).
func FakeTLEResult(totalCases, failIndex int) *JudgeResult {
	cases := make([]Verdict, totalCases)
	for i := range cases {
		if i < failIndex {
			cases[i] = Verdict{Status: "AC", TimeMs: 10, MemoryKb: 1024}
		} else if i == failIndex {
			cases[i] = Verdict{Status: "TLE", TimeMs: 1000, MemoryKb: 1024}
		} else {
			cases[i] = Verdict{Status: "TLE", TimeMs: 0, MemoryKb: 0}
		}
	}
	return &JudgeResult{
		Status:       "TLE",
		PassedCases:  failIndex,
		TotalCases:   totalCases,
		TotalTimeMs:  failIndex*10 + 1000,
		PeakMemoryKb: 1024,
		CaseResults:  cases,
	}
}

// FakeMLEResult returns a mock MLE result.
func FakeMLEResult(totalCases, failIndex int) *JudgeResult {
	cases := make([]Verdict, totalCases)
	for i := range cases {
		if i < failIndex {
			cases[i] = Verdict{Status: "AC", TimeMs: 10, MemoryKb: 1024}
		} else if i == failIndex {
			cases[i] = Verdict{Status: "MLE", TimeMs: 10, MemoryKb: 262144}
		} else {
			cases[i] = Verdict{Status: "MLE", TimeMs: 0, MemoryKb: 0}
		}
	}
	return &JudgeResult{
		Status:       "MLE",
		PassedCases:  failIndex,
		TotalCases:   totalCases,
		TotalTimeMs:  failIndex*10 + 10,
		PeakMemoryKb: 262144,
		CaseResults:  cases,
	}
}

// FakeREResult returns a mock RE result.
func FakeREResult(totalCases, failIndex int) *JudgeResult {
	cases := make([]Verdict, totalCases)
	for i := range cases {
		if i < failIndex {
			cases[i] = Verdict{Status: "AC", TimeMs: 10, MemoryKb: 1024}
		} else if i == failIndex {
			cases[i] = Verdict{Status: "RE", TimeMs: 10, MemoryKb: 1024, ActualOutput: "Runtime Error (SIGSEGV)"}
		} else {
			cases[i] = Verdict{Status: "RE", TimeMs: 0, MemoryKb: 0}
		}
	}
	return &JudgeResult{
		Status:       "RE",
		PassedCases:  failIndex,
		TotalCases:   totalCases,
		TotalTimeMs:  failIndex*10 + 10,
		PeakMemoryKb: 1024,
		CaseResults:  cases,
	}
}

// FakeSystemErrorResult returns a mock SYSTEM_ERROR result.
func FakeSystemErrorResult(message string) *JudgeResult {
	return &JudgeResult{
		Status: "SYSTEM_ERROR",
		CaseResults: []Verdict{
			{Status: "SYSTEM_ERROR", ActualOutput: message},
		},
	}
}
