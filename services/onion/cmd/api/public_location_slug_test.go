package main

import "testing"

func TestPublicLocationSlug(t *testing.T) {
	cases := []struct {
		name string
		id   int64
		want string
	}{
		{"Voxire Cafe", 4, "voxire-cafe"},
		{"  Downtown — #1! ", 9, "downtown-1"},
		{"فرع بيروت", 7, "branch-7"},
		{"", 12, "branch-12"},
	}
	for _, tc := range cases {
		if got := publicLocationSlug(tc.name, tc.id); got != tc.want {
			t.Errorf("publicLocationSlug(%q, %d) = %q, want %q", tc.name, tc.id, got, tc.want)
		}
	}
}
