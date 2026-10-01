package services

import (
	"errors"
	"testing"

	"github.com/zechtz/vertex/internal/models"
)

const testPassword = "correct-horse-battery"

func assertCanLogin(t *testing.T, as *AuthService, password string) {
	t.Helper()

	if _, err := as.Login(&models.UserLogin{Email: "tester@example.com", Password: password}); err != nil {
		t.Errorf("login with %q failed: %v", password, err)
	}
}

func assertCannotLogin(t *testing.T, as *AuthService, password string) {
	t.Helper()

	if _, err := as.Login(&models.UserLogin{Email: "tester@example.com", Password: password}); err == nil {
		t.Errorf("login with %q succeeded, want it refused", password)
	}
}

func TestChangePasswordReplacesThePassword(t *testing.T) {
	as := newTestAuthService(t)
	user := registerTestUser(t, as)

	err := as.ChangePassword(user.ID, &models.PasswordChange{
		CurrentPassword: testPassword,
		NewPassword:     "a-new-password",
	})
	if err != nil {
		t.Fatalf("ChangePassword returned error: %v", err)
	}

	assertCanLogin(t, as, "a-new-password")
	assertCannotLogin(t, as, testPassword)
}

// A session alone must not be enough to take over an account.
func TestChangePasswordRequiresTheCurrentPassword(t *testing.T) {
	as := newTestAuthService(t)
	user := registerTestUser(t, as)

	err := as.ChangePassword(user.ID, &models.PasswordChange{
		CurrentPassword: "not-the-password",
		NewPassword:     "a-new-password",
	})
	if !errors.Is(err, ErrIncorrectPassword) {
		t.Fatalf("ChangePassword error = %v, want ErrIncorrectPassword", err)
	}

	assertCanLogin(t, as, testPassword)
}

func TestChangePasswordRejectsAPasswordRegistrationWouldRefuse(t *testing.T) {
	as := newTestAuthService(t)
	user := registerTestUser(t, as)

	err := as.ChangePassword(user.ID, &models.PasswordChange{
		CurrentPassword: testPassword,
		NewPassword:     "short",
	})
	if err == nil {
		t.Fatal("ChangePassword accepted a 5-character password")
	}

	assertCanLogin(t, as, testPassword)
}

func TestResetPasswordNeedsNoCurrentPassword(t *testing.T) {
	as := newTestAuthService(t)
	registerTestUser(t, as)

	user, err := as.ResetPassword("tester@example.com", "a-new-password")
	if err != nil {
		t.Fatalf("ResetPassword returned error: %v", err)
	}
	if user.Password != "" {
		t.Error("ResetPassword returned the password hash")
	}

	assertCanLogin(t, as, "a-new-password")
	assertCannotLogin(t, as, testPassword)
}

func TestResetPasswordReportsAnUnknownEmail(t *testing.T) {
	as := newTestAuthService(t)
	registerTestUser(t, as)

	_, err := as.ResetPassword("someone@example.com", "a-new-password")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("ResetPassword error = %v, want ErrUserNotFound", err)
	}

	assertCanLogin(t, as, testPassword)
}

func TestListUsersOmitsPasswordHashes(t *testing.T) {
	as := newTestAuthService(t)
	registerTestUser(t, as)

	users, err := as.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers returned error: %v", err)
	}

	if len(users) != 1 || users[0].Email != "tester@example.com" {
		t.Fatalf("ListUsers = %+v, want the one registered user", users)
	}
	if users[0].Password != "" {
		t.Error("ListUsers returned a password hash")
	}
}
