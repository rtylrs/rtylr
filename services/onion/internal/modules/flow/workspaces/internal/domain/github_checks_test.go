package domain

import "testing"

func TestChecksState(t *testing.T) {
	ok := GitHubCheck{Status: "completed", Conclusion: "success"}
	skipped := GitHubCheck{Status: "completed", Conclusion: "skipped"}
	running := GitHubCheck{Status: "in_progress"}
	failed := GitHubCheck{Status: "completed", Conclusion: "failure"}
	cases := []struct {
		checks []GitHubCheck
		want   string
	}{
		{nil, ChecksNone},
		{[]GitHubCheck{ok, skipped}, ChecksPassing},
		{[]GitHubCheck{ok, running}, ChecksRunning},
		{[]GitHubCheck{running, failed, ok}, ChecksFailing},
		{[]GitHubCheck{{Status: "completed", Conclusion: "timed_out"}}, ChecksFailing},
	}
	for i, c := range cases {
		if got := ChecksState(c.checks); got != c.want {
			t.Errorf("case %d: %s, want %s", i, got, c.want)
		}
	}
	if WorseChecksState(ChecksPassing, ChecksRunning) != ChecksRunning || WorseChecksState(ChecksFailing, ChecksNone) != ChecksFailing {
		t.Error("WorseChecksState picks the wrong state")
	}
}
