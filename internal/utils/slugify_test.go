package utils

import "testing"

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple case",
			input: "Hello World",
			want:  "hello-world",
		},
		{
			name:  "with extra spaces",
			input: "  leading and trailing spaces  ",
			want:  "leading-and-trailing-spaces",
		},
		{
			name:  "with special characters",
			input: "special!@#$%^&*()_+-=[]{}|;':,./<>?`~characters",
			want:  "special-characters",
		},
		{
			name:  "with mixed case",
			input: "MixedCase Test",
			want:  "mixedcase-test",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "string with only special characters",
			input: "!@#$%^&*()",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Slugify(tt.input); got != tt.want {
				t.Errorf("Slugify() = %v, want %v", got, tt.want)
			}
		})
	}
}
