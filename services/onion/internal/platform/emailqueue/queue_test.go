package emailqueue

import "testing"

func TestRetryDelayBackoff(t *testing.T) {
	tests := []struct {
		attempt int
		want    int
	}{
		{attempt: 1, want: 30},
		{attempt: 2, want: 120},
		{attempt: 3, want: 300},
		{attempt: 99, want: 300},
	}
	for _, tt := range tests {
		if got := retryDelay(tt.attempt); got != tt.want {
			t.Errorf("retryDelay(%d) = %d, want %d", tt.attempt, got, tt.want)
		}
	}
}
