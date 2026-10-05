package infrastructure

import (
	"database/sql"
	"errors"
	"testing"
)

type fakeStatusUpdateResult struct {
	affected int64
	err      error
}

func (r fakeStatusUpdateResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (r fakeStatusUpdateResult) RowsAffected() (int64, error) {
	return r.affected, r.err
}

func TestStatusUpdateChanged(t *testing.T) {
	tests := []struct {
		name    string
		result  sql.Result
		want    bool
		wantErr bool
	}{
		{name: "changed", result: fakeStatusUpdateResult{affected: 1}, want: true},
		{name: "unchanged", result: fakeStatusUpdateResult{}, want: false},
		{name: "rows affected error", result: fakeStatusUpdateResult{err: errors.New("rows affected failed")}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := statusUpdateChanged(tt.result)
			if (err != nil) != tt.wantErr {
				t.Fatalf("statusUpdateChanged() error = %v, wantErr %t", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("statusUpdateChanged() = %t, want %t", got, tt.want)
			}
		})
	}
}
