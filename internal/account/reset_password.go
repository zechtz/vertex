// Package account holds the command-line operations on Vertex user accounts.
package account

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/zechtz/vertex/internal/database"
	"github.com/zechtz/vertex/internal/models"
	"github.com/zechtz/vertex/internal/services"
	"golang.org/x/term"
)

// ResetPassword sets a new password for the account with this email, for a
// user who has forgotten theirs. Being able to run Vertex against its database
// is what proves the right to do it, so it needs neither the old password nor
// email delivery.
//
// At a terminal the new password is typed, twice and unechoed. Without one -
// a script, or docker exec without -t - a random password is generated and
// printed once, to be changed after signing in.
func ResetPassword(email string) error {
	// Opening the database and the auth service log startup chatter meant for
	// the server's log, not for someone at a prompt. Failures still come back
	// as errors.
	log.SetOutput(io.Discard)

	db, err := database.NewDatabase()
	if err != nil {
		return fmt.Errorf("failed to open the database: %w", err)
	}
	defer db.Close()

	auth := services.NewAuthService(db)

	users, err := auth.ListUsers()
	if err != nil {
		return err
	}

	if !hasAccount(users, email) {
		return noAccountError(email, users, db.Path())
	}

	interactive := term.IsTerminal(int(os.Stdin.Fd()))

	var password string
	if interactive {
		password, err = promptNewPassword()
	} else {
		password, err = generatePassword()
	}
	if err != nil {
		return err
	}

	user, err := auth.ResetPassword(email, password)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Password reset for %s (%s)\n", user.Username, user.Email)
	if !interactive {
		fmt.Printf("\nTemporary password: %s\n\n", password)
		fmt.Println("Sign in with it, then change it from the user menu (Change Password).")
	}

	return nil
}

func hasAccount(users []models.User, email string) bool {
	for _, user := range users {
		if user.Email == email {
			return true
		}
	}
	return false
}

// noAccountError names the database it looked in and the accounts that do
// exist: a mistyped email and Vertex looking at the wrong data directory are
// the two likely causes, and this tells them apart.
func noAccountError(email string, users []models.User, dbPath string) error {
	if len(users) == 0 {
		return fmt.Errorf("there are no accounts in %s\n"+
			"If Vertex keeps its data elsewhere, pass --data-dir or set VERTEX_DATA_DIR", dbPath)
	}

	message := fmt.Sprintf("no account with email %q in %s\nAccounts:", email, dbPath)
	for _, user := range users {
		message += fmt.Sprintf("\n  %s (%s)", user.Email, user.Username)
	}

	return fmt.Errorf("%s", message)
}

func promptNewPassword() (string, error) {
	password, err := readPassword("New password: ")
	if err != nil {
		return "", err
	}

	if err := models.ValidatePassword(password); err != nil {
		return "", err
	}

	confirmation, err := readPassword("Confirm new password: ")
	if err != nil {
		return "", err
	}

	if password != confirmation {
		return "", fmt.Errorf("passwords do not match")
	}

	return password, nil
}

func readPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	return string(password), nil
}

// generatePassword returns 24 random URL-safe characters.
func generatePassword() (string, error) {
	bytes := make([]byte, 18)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate a password: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
