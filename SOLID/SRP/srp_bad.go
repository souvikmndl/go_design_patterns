package srp

import (
	"errors"
	"fmt"
)

// UserManager handles everything. This is an anti-pattern.
type UserManager struct{}

// RegisterUser is saving to db and sending notifications
// However, all that dependency is on the UserManager struct
// nothing can be mocked, if you want to change DBs or Email sender
// you will have to individually change everything connected to the
// UserManager, chances of making mistakes etc
func (um *UserManager) RegisterUser(u User) error {
	// Responsibility 1: Business Logic / Validation
	if u.Email == "" {
		return errors.New("email cannot be empty")
	}

	// Responsibility 2: Database Persistence
	fmt.Printf("Saving %s to the database...\n", u.Name)
	// e.g., db.Exec("INSERT INTO users...")

	// Responsibility 3: Notifications
	fmt.Printf("Sending welcome email to %s...\n", u.Email)
	// e.g., smtp.SendMail(...)

	return nil
}
