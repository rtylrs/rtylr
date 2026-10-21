package domain

import "testing"

func TestPeriodHasPaidRecord(t *testing.T) {
	tests := []struct {
		name string
		rows []PeriodRow
		want bool
	}{
		{
			name: "empty period is not paid",
			rows: nil,
			want: false,
		},
		{
			name: "all drafts are not paid",
			rows: []PeriodRow{{ID: 1, StaffID: 4101, Status: "draft"}, {ID: 2, StaffID: 4102, Status: "processed"}},
			want: false,
		},
		{
			name: "one settled employee makes the period paid",
			rows: []PeriodRow{{ID: 1, StaffID: 4101, Status: "draft"}, {ID: 2, StaffID: 4102, Status: StatusPaid}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PeriodHasPaidRecord(tt.rows); got != tt.want {
				t.Errorf("PeriodHasPaidRecord() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStaffHasPaidRecordInPeriod(t *testing.T) {
	tests := []struct {
		name     string
		rows     []PeriodRow
		staffID  int64
		exceptID int64
		want     bool
	}{
		{
			name:    "no rows at all",
			rows:    nil,
			staffID: 4101,
			want:    false,
		},
		{
			name:     "settled sibling for the same employee is a duplicate",
			rows:     []PeriodRow{{ID: 1, StaffID: 4101, Status: StatusPaid}, {ID: 2, StaffID: 4101, Status: "draft"}},
			staffID:  4101,
			exceptID: 2,
			want:     true,
		},
		{
			name:     "the record's own paid status is not a sibling",
			rows:     []PeriodRow{{ID: 2, StaffID: 4101, Status: StatusPaid}},
			staffID:  4101,
			exceptID: 2,
			want:     false,
		},
		{
			// Paying a period out employee by employee is the normal path, so a
			// settled record for a different employee must not block this one.
			name:     "settled record for a different employee is not a conflict",
			rows:     []PeriodRow{{ID: 1, StaffID: 4102, Status: StatusPaid}, {ID: 2, StaffID: 4101, Status: "draft"}},
			staffID:  4101,
			exceptID: 2,
			want:     false,
		},
		{
			name:     "unsettled duplicate is not yet a conflict",
			rows:     []PeriodRow{{ID: 1, StaffID: 4101, Status: "draft"}, {ID: 2, StaffID: 4101, Status: "draft"}},
			staffID:  4101,
			exceptID: 2,
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StaffHasPaidRecordInPeriod(tt.rows, tt.staffID, tt.exceptID); got != tt.want {
				t.Errorf("StaffHasPaidRecordInPeriod() = %v, want %v", got, tt.want)
			}
		})
	}
}
