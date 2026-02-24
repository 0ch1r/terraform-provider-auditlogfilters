package provider

import "testing"

func TestParseNonNegativeInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		want      int
		expectErr bool
	}{
		{
			name:      "valid_zero",
			input:     "0",
			want:      0,
			expectErr: false,
		},
		{
			name:      "valid_positive",
			input:     "7",
			want:      7,
			expectErr: false,
		},
		{
			name:      "negative",
			input:     "-1",
			expectErr: true,
		},
		{
			name:      "non_integer",
			input:     "abc",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseNonNegativeInt(tc.input)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error for input %q", tc.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tc.input, err)
			}

			if got != tc.want {
				t.Fatalf("unexpected value for input %q: got %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}
