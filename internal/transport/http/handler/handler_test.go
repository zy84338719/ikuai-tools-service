package handler

import "testing"

func TestParseIDs(t *testing.T) {
	tests := []struct {
		in     string
		want   []int
		hasErr bool
	}{
		{"", nil, false},
		{"1", []int{1}, false},
		{"1,2,3", []int{1, 2, 3}, false},
		{" 1 , 2 ", []int{1, 2}, false},
		{"1,,2", []int{1, 2}, false},
		{"abc", nil, true},
		{"1,x", nil, true},
	}
	for _, tt := range tests {
		got, err := parseIDs(tt.in)
		if (err != nil) != tt.hasErr {
			t.Errorf("parseIDs(%q) err=%v, want hasErr=%v", tt.in, err, tt.hasErr)
			continue
		}
		if !equalInts(got, tt.want) {
			t.Errorf("parseIDs(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestJoinInts(t *testing.T) {
	tests := []struct {
		in   []int
		want string
	}{
		{nil, ""},
		{[]int{1}, "1"},
		{[]int{1, 2, 3}, "1,2,3"},
	}
	for _, tt := range tests {
		if got := joinInts(tt.in); got != tt.want {
			t.Errorf("joinInts(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
