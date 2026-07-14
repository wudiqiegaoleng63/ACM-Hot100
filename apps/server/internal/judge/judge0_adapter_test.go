package judge

import (
	"testing"
)

func TestMapJudge0StatusMapsAllKnownStatuses(t *testing.T) {
	cases := []struct {
		id   int
		want string
	}{
		{3, "AC"},
		{4, "WA"},
		{5, "TLE"},
		{6, "MLE"},
		{7, "RE"},
		{8, "RE"},
		{9, "RE"},
		{13, "CE"},
		{10, "SYSTEM_ERROR"},
		{11, "SYSTEM_ERROR"},
		{99, "SYSTEM_ERROR"}, // unknown
	}
	for _, tc := range cases {
		got := mapJudge0Status(tc.id)
		if got != tc.want {
			t.Errorf("mapJudge0Status(%d) = %s, want %s", tc.id, got, tc.want)
		}
	}
}

func TestFakeRuntimeAndSystemErrorsExposeOnlySanitizedSummaries(t *testing.T) {
	runtimeResult := FakeREResult(3, 1)
	if runtimeResult.ErrorMessage != "Runtime Error (SIGSEGV)" {
		t.Fatalf("runtime error message = %q", runtimeResult.ErrorMessage)
	}

	systemResult := FakeSystemErrorResult("failed at /home/judge/private/main.cpp")
	if systemResult.ErrorMessage == "failed at /home/judge/private/main.cpp" {
		t.Fatal("system error must not preserve the host path")
	}
	if systemResult.CaseResults[0].ActualOutput != "" {
		t.Fatal("system error details must not be stored as hidden-case output")
	}
}

func TestParseTimeMs(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"0.045", 45},
		{"1.500", 1500},
		{"0", 0},
		{"", 0},
	}
	for _, tc := range cases {
		got := parseTimeMs(tc.input)
		if got != tc.want {
			t.Errorf("parseTimeMs(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}
