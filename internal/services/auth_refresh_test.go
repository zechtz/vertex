package services

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/zechtz/vertex/internal/database"
	"github.com/zechtz/vertex/internal/models"
)

func newTestAuthService(t *testing.T) *AuthService {
	t.Helper()

	db, err := database.NewDatabaseWithPath(filepath.Join(t.TempDir(), "vertex.db"))
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return NewAuthService(db)
}

func registerTestUser(t *testing.T, as *AuthService) *models.User {
	t.Helper()

	user, err := as.Register(&models.UserRegistration{
		Username: "tester",
		Email:    "tester@example.com",
		Password: "correct-horse-battery",
	})
	if err != nil {
		t.Fatalf("failed to register user: %v", err)
	}

	return user
}

// Refreshing must hand back a usable session for the same user, so the client
// can swap the old token for the new one without asking for a password.
func TestRefreshTokenIssuesAUsableSession(t *testing.T) {
	as := newTestAuthService(t)
	user := registerTestUser(t, as)

	refreshed, err := as.RefreshToken(user.ID)
	if err != nil {
		t.Fatalf("RefreshToken returned error: %v", err)
	}

	claims, err := as.ValidateToken(refreshed.Token)
	if err != nil {
		t.Fatalf("refreshed token did not validate: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("refreshed token is for user %q, want %q", claims.UserID, user.ID)
	}

	// The point of refreshing is a later expiry than the token being replaced.
	if claims.ExpiresAt == nil {
		t.Fatal("refreshed token has no expiry")
	}
	if remaining := time.Until(claims.ExpiresAt.Time); remaining < 23*time.Hour {
		t.Errorf("refreshed token expires in %v, want a full lifetime", remaining)
	}

	// The password hash must never ride along in a response.
	if refreshed.User.Password != "" {
		t.Error("refresh response contains the password hash")
	}
}

// A token naming a user who no longer exists must not be renewable, otherwise a
// deleted account keeps working until its last token happens to lapse.
func TestRefreshTokenRejectsUnknownUser(t *testing.T) {
	as := newTestAuthService(t)

	if _, err := as.RefreshToken("no-such-user"); err == nil {
		t.Error("RefreshToken succeeded for a user that does not exist")
	}
}
