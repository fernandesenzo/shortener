package identity_test

import (
	"context"
	"testing"

	"github.com/fernandesenzo/shortener/internal/identity"
)

type contextKey string

const otherKey contextKey = "otherKey"

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		wantID string
		wantOK bool
	}{
		{
			name:   "userID set",
			ctx:    identity.WithUserID(context.Background(), "1234"),
			wantID: "1234",
			wantOK: true,
		},
		{
			name:   "userID not set",
			ctx:    context.Background(),
			wantID: "",
			wantOK: false,
		},
		{
			name:   "different context value should not interfere",
			ctx:    context.WithValue(context.Background(), otherKey, "value"),
			wantID: "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := identity.GetUserID(tt.ctx)

			if ok != tt.wantOK {
				t.Fatalf("expected ok=%v, got %v", tt.wantOK, ok)
			}

			if id != tt.wantID {
				t.Fatalf("expected id=%q, got %q", tt.wantID, id)
			}
		})
	}
}
