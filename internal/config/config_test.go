package config

import "testing"

func TestDefaultString(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		fallback string
		want     string
	}{
		{name: "use input", in: "value", fallback: "fb", want: "value"},
		{name: "trim input", in: " value ", fallback: "fb", want: "value"},
		{name: "fallback when empty", in: "", fallback: "fb", want: "fb"},
		{name: "fallback when spaces", in: "   ", fallback: "fb", want: "fb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultString(tt.in, tt.fallback); got != tt.want {
				t.Fatalf("defaultString(%q,%q)=%q want=%q", tt.in, tt.fallback, got, tt.want)
			}
		})
	}
}
