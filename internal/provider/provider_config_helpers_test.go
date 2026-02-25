package provider

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestParsePositiveInt64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		want      int64
		wantErr   bool
		wantRange bool
	}{
		{name: "valid", input: "15", want: 15},
		{name: "zero", input: "0", wantErr: true, wantRange: true},
		{name: "negative", input: "-1", wantErr: true, wantRange: true},
		{name: "non_integer", input: "abc", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parsePositiveInt64(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q", tc.input)
				}
				if tc.wantRange && !errors.Is(err, errNonPositiveInt64) {
					t.Fatalf("expected range validation error for input %q, got: %v", tc.input, err)
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

func TestValidatePositiveInt64(t *testing.T) {
	t.Parallel()

	if err := validatePositiveInt64(1); err != nil {
		t.Fatalf("unexpected error for positive value: %v", err)
	}

	err := validatePositiveInt64(0)
	if !errors.Is(err, errNonPositiveInt64) {
		t.Fatalf("expected non-positive validation error for zero, got: %v", err)
	}
}

func TestConfigStringOrEnv(t *testing.T) {
	t.Parallel()

	envValue := "from-env"
	cfgValue := "from-config"

	got := configStringOrEnv(types.StringValue(cfgValue), envValue)
	if got != cfgValue {
		t.Fatalf("expected config value precedence, got: %q", got)
	}

	got = configStringOrEnv(types.StringNull(), envValue)
	if got != envValue {
		t.Fatalf("expected env fallback, got: %q", got)
	}
}
