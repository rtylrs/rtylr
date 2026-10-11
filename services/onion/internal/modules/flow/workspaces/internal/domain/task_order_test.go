package domain_test

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

func TestValidateTaskOrder(t *testing.T) {
	tooMany := make([]int64, domain.MaxOrderedTaskIDs+1)
	for i := range tooMany {
		tooMany[i] = int64(i + 1)
	}
	atLimit := tooMany[:domain.MaxOrderedTaskIDs]

	cases := []struct {
		name    string
		moved   int64
		ids     []int64
		wantErr bool
	}{
		{"empty is legacy", 5, nil, false},
		{"valid", 5, []int64{3, 5, 9}, false},
		{"exactly at the cap", 1, atLimit, false},
		{"over the cap", 1, tooMany, true},
		{"zero id", 5, []int64{5, 0}, true},
		{"negative id", 5, []int64{5, -2}, true},
		{"duplicate", 5, []int64{5, 3, 3}, true},
		{"moved task missing", 5, []int64{1, 2}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidateTaskOrder(tc.moved, tc.ids)
			if tc.wantErr {
				if !errors.Is(err, domain.ErrInvalidTaskOrder) {
					t.Fatalf("err = %v, want ErrInvalidTaskOrder", err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
