package groups

import (
	"errors"
	"testing"
)

func TestCreateGroupRequestValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		req     CreateGroupRequest
		wantErr bool
	}{
		{name: "valid", req: CreateGroupRequest{Email: "team@example.com"}, wantErr: false},
		{name: "valid with name + description", req: CreateGroupRequest{Email: "x@example.com", Name: "Team", Description: "All"}, wantErr: false},
		{name: "missing email", req: CreateGroupRequest{}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.req.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate() err = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.wantErr && !IsInvalid(err) {
				t.Errorf("expected validation error to satisfy IsInvalid, got %v", err)
			}
		})
	}
}

func TestAddMemberRequestValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		req     AddMemberRequest
		wantErr bool
	}{
		{name: "valid", req: AddMemberRequest{Email: "a@example.com"}, wantErr: false},
		{name: "valid with role", req: AddMemberRequest{Email: "a@example.com", Role: "OWNER"}, wantErr: false},
		{name: "missing email", req: AddMemberRequest{}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.req.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate() err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestIsInvalid(t *testing.T) {
	t.Parallel()
	if !IsInvalid(ErrInvalid("x")) {
		t.Error("expected IsInvalid(ErrInvalid) to be true")
	}
	if IsInvalid(nil) {
		t.Error("expected IsInvalid(nil) to be false")
	}
	if IsInvalid(errors.New("plain")) {
		t.Error("expected IsInvalid(plain error) to be false")
	}
}
