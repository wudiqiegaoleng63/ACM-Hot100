package judge

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCompareOutputMatchesIdenticalContent(t *testing.T) {
	if !CompareOutput("1 2 3\n", "1 2 3\n") {
		t.Fatal("identical content should match")
	}
}

func TestCompareOutputNormalizesCRLF(t *testing.T) {
	if !CompareOutput("1 2 3\r\n", "1 2 3\n") {
		t.Fatal("CRLF should be normalized to LF")
	}
}

func TestCompareOutputNormalizesCR(t *testing.T) {
	if !CompareOutput("1 2 3\r", "1 2 3\n") {
		t.Fatal("CR should be normalized to LF")
	}
}

func TestCompareOutputTrimsTrailingWhitespace(t *testing.T) {
	if !CompareOutput("1 2 3  \n", "1 2 3\n") {
		t.Fatal("trailing whitespace should be trimmed")
	}
}

func TestCompareOutputRemovesTrailingBlankLines(t *testing.T) {
	if !CompareOutput("1 2 3\n\n\n", "1 2 3\n") {
		t.Fatal("trailing blank lines should be removed")
	}
}

func TestCompareOutputRejectsDifferentContent(t *testing.T) {
	if CompareOutput("1 2 3\n", "1 2 4\n") {
		t.Fatal("different content should not match")
	}
}

func TestCompareOutputIsCaseSensitive(t *testing.T) {
	if CompareOutput("abc\n", "ABC\n") {
		t.Fatal("comparison should be case-sensitive")
	}
}

func TestCompareOutputComplexNormalization(t *testing.T) {
	actual := "  hello  \r\nworld  \r\n\r\n"
	expected := "  hello\nworld\n"
	if !CompareOutput(actual, expected) {
		t.Fatal("complex normalization should match")
	}
}

func TestTruncateOutputPreservesShortOutput(t *testing.T) {
	short := "hello"
	if TruncateOutput(short) != short {
		t.Fatal("short output should be preserved")
	}
}

func TestTruncateOutputTruncatesLongOutput(t *testing.T) {
	long := strings.Repeat("x", 10*1024)
	result := TruncateOutput(long)
	if len(result) > maxOutputBytes+50 { // some slack for the marker
		t.Fatalf("truncated output length = %d, should be near %d", len(result), maxOutputBytes)
	}
	if !strings.Contains(result, "[truncated]") {
		t.Fatal("truncated output should contain marker")
	}
}

func TestSanitizePathRemovesHostPaths(t *testing.T) {
	input := "Error at /home/user/code/main.cpp:5"
	result := SanitizePath(input)
	if strings.Contains(result, "/home/") {
		t.Fatalf("host path should be removed, got: %s", result)
	}
	if !strings.Contains(result, "[path]") {
		t.Fatal("host path should be replaced with [path]")
	}
}

func TestFakeAdapterReturnsConfiguredResult(t *testing.T) {
	expected := FakeACResult(3)
	adapter := NewFakeAdapter(expected)
	result, err := adapter.Judge(context.Background(), "sub-1")
	if err != nil {
		t.Fatalf("Judge: %v", err)
	}
	if result.Status != "AC" {
		t.Fatalf("status = %s, want AC", result.Status)
	}
	if result.PassedCases != 3 {
		t.Fatalf("passed = %d, want 3", result.PassedCases)
	}
}

func TestFakeAdapterReturnsError(t *testing.T) {
	adapter := NewFakeAdapterWithError(errors.New("connection refused"))
	_, err := adapter.Judge(context.Background(), "sub-1")
	if err == nil {
		t.Fatal("should return configured error")
	}
}

func TestFakeWAResultStopsAtFailIndex(t *testing.T) {
	result := FakeWAResult(5, 2)
	if result.Status != "WA" {
		t.Fatalf("status = %s, want WA", result.Status)
	}
	if result.PassedCases != 2 {
		t.Fatalf("passed = %d, want 2", result.PassedCases)
	}
	if result.CaseResults[0].Status != "AC" {
		t.Fatalf("case 0 = %s, want AC", result.CaseResults[0].Status)
	}
	if result.CaseResults[2].Status != "WA" {
		t.Fatalf("case 2 = %s, want WA", result.CaseResults[2].Status)
	}
	// After the failing case, remaining should also be WA (not executed)
	if result.CaseResults[3].Status != "WA" {
		t.Fatalf("case 3 = %s, want WA (not executed)", result.CaseResults[3].Status)
	}
}

func TestFakeCEResultHasNoTestCases(t *testing.T) {
	result := FakeCEResult()
	if result.Status != "CE" {
		t.Fatalf("status = %s, want CE", result.Status)
	}
	if result.TotalCases != 0 {
		t.Fatalf("total = %d, want 0 for CE", result.TotalCases)
	}
	if result.CompilerOutput == "" {
		t.Fatal("CE should have compiler output")
	}
}

func TestFakeTLEResult(t *testing.T) {
	result := FakeTLEResult(3, 1)
	if result.Status != "TLE" {
		t.Fatalf("status = %s, want TLE", result.Status)
	}
	if result.PassedCases != 1 {
		t.Fatalf("passed = %d, want 1", result.PassedCases)
	}
}

func TestFakeMLEResult(t *testing.T) {
	result := FakeMLEResult(3, 0)
	if result.Status != "MLE" {
		t.Fatalf("status = %s, want MLE", result.Status)
	}
	if result.PassedCases != 0 {
		t.Fatalf("passed = %d, want 0", result.PassedCases)
	}
}

func TestFakeREResult(t *testing.T) {
	result := FakeREResult(4, 3)
	if result.Status != "RE" {
		t.Fatalf("status = %s, want RE", result.Status)
	}
	if result.PassedCases != 3 {
		t.Fatalf("passed = %d, want 3", result.PassedCases)
	}
}

func TestFakeSystemErrorResult(t *testing.T) {
	result := FakeSystemErrorResult("judge timeout")
	if result.Status != "SYSTEM_ERROR" {
		t.Fatalf("status = %s, want SYSTEM_ERROR", result.Status)
	}
}

func TestIsTerminalSubmissionStatus(t *testing.T) {
	terminal := []string{"AC", "WA", "TLE", "MLE", "RE", "CE", "SYSTEM_ERROR"}
	for _, s := range terminal {
		if !IsTerminalSubmissionStatus(s) {
			t.Fatalf("%s should be terminal", s)
		}
	}
	nonTerminal := []string{"QUEUED", "COMPILING", "RUNNING"}
	for _, s := range nonTerminal {
		if IsTerminalSubmissionStatus(s) {
			t.Fatalf("%s should not be terminal", s)
		}
	}
}
