package bot

import "testing"

func TestIsOptionAnswer(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"A", true},
		{"b", true},
		{"C) option", true},
		{"d answer", true},
		{"", false},
		{"x", false},
		{"1", false},
	}

	for _, tc := range cases {
		if got := isOptionAnswer(tc.in); got != tc.want {
			t.Fatalf("isOptionAnswer(%q)=%v want=%v", tc.in, got, tc.want)
		}
	}
}
