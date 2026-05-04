package users

import (
	"errors"
	"testing"
)

func TestCreateRequestValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		req     CreateRequest
		wantErr bool
		invalid bool
	}{
		{
			name:    "valid",
			req:     CreateRequest{PrimaryEmail: "a@example.com", GivenName: "A", FamilyName: "B", Password: "longenough"},
			wantErr: false,
		},
		{
			name:    "missing email",
			req:     CreateRequest{GivenName: "A", FamilyName: "B", Password: "longenough"},
			wantErr: true,
			invalid: true,
		},
		{
			name:    "missing given name",
			req:     CreateRequest{PrimaryEmail: "a@example.com", FamilyName: "B", Password: "longenough"},
			wantErr: true,
			invalid: true,
		},
		{
			name:    "missing family name",
			req:     CreateRequest{PrimaryEmail: "a@example.com", GivenName: "A", Password: "longenough"},
			wantErr: true,
			invalid: true,
		},
		{
			name:    "short password",
			req:     CreateRequest{PrimaryEmail: "a@example.com", GivenName: "A", FamilyName: "B", Password: "short"},
			wantErr: true,
			invalid: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.req.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate() err = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.invalid && !IsInvalid(err) {
				t.Errorf("expected IsInvalid(err) to be true, got err = %v", err)
			}
		})
	}
}

func TestAddAliasRequestValidate(t *testing.T) {
	t.Parallel()
	if err := (AddAliasRequest{Alias: "a@example.com"}).Validate(); err != nil {
		t.Errorf("expected valid alias, got %v", err)
	}
	err := (AddAliasRequest{}).Validate()
	if err == nil {
		t.Fatal("expected error for empty alias")
	}
	if !IsInvalid(err) {
		t.Errorf("expected IsInvalid(err) for empty alias, got %v", err)
	}
}

func TestIsInvalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "ErrInvalid", err: ErrInvalid("bad"), want: true},
		{name: "stdlib error", err: errors.New("oops"), want: false},
		{name: "ErrWorkspaceUnavailable", err: ErrWorkspaceUnavailable, want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := IsInvalid(tc.err); got != tc.want {
				t.Errorf("IsInvalid(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
