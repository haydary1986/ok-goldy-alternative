package config

import (
	"reflect"
	"testing"
)

func TestSplitCSV(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{name: "empty", in: "", want: nil},
		{name: "single", in: "a@example.com", want: []string{"a@example.com"}},
		{name: "multi", in: "a@example.com,b@example.com,c@example.com", want: []string{"a@example.com", "b@example.com", "c@example.com"}},
		{name: "spaces trimmed", in: " a@example.com , b@example.com ", want: []string{"a@example.com", "b@example.com"}},
		{name: "empty entries dropped", in: "a@example.com,,b@example.com", want: []string{"a@example.com", "b@example.com"}},
		{name: "all empty", in: ",,, ", want: nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := splitCSV(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("splitCSV(%q) = %#v, want %#v", tc.in, got, tc.want)
			}
		})
	}
}

// NOTE: TestGetenv and TestGetenvInt do not call t.Parallel because they use
// t.Setenv, which mutates process-global state and is incompatible with
// parallel execution.

func TestGetenv(t *testing.T) {
	const key = "GOLDY_TEST_GETENV_KEY"

	t.Run("unset returns default", func(t *testing.T) {
		t.Setenv(key, "")
		if got := getenv(key, "fallback"); got != "fallback" {
			t.Errorf("getenv unset = %q, want %q", got, "fallback")
		}
	})

	t.Run("set returns value", func(t *testing.T) {
		t.Setenv(key, "explicit")
		if got := getenv(key, "fallback"); got != "explicit" {
			t.Errorf("getenv set = %q, want %q", got, "explicit")
		}
	})
}

func TestGetenvInt(t *testing.T) {
	const key = "GOLDY_TEST_GETENV_INT"

	tests := []struct {
		name string
		val  string
		def  int
		want int
	}{
		{name: "empty -> default", val: "", def: 42, want: 42},
		{name: "valid int", val: "100", def: 0, want: 100},
		{name: "non-numeric -> default", val: "twenty", def: 7, want: 7},
		{name: "negative int", val: "-1", def: 0, want: -1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(key, tc.val)
			if got := getenvInt(key, tc.def); got != tc.want {
				t.Errorf("getenvInt(%q, %d) = %d, want %d", tc.val, tc.def, got, tc.want)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{name: "missing database URL", cfg: Config{}, wantErr: true},
		{name: "minimum valid", cfg: Config{DatabaseURL: "postgres://x"}, wantErr: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.cfg.validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("validate() err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}
