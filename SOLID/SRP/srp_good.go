package srp

import (
	"errors"
	"fmt"
)

// 1. Core domain model
// User is the core domain model
type User struct {
	Name  string
	Email string
}

// 2. Responsibility: DB Persistence
// UserRepository is DB persistence layer
type UserRepository interface {
	Save(u User) error
}

// PSQLUserRepository is the actual db implementation
type PSQLUserRepository struct{}

// Save implements UserRepository's Save() method
func (repo *PSQLUserRepository) Save(u User) error {
	fmt.Printf("Saving %s to PSQL db...\n", u.Name)
	return nil
}

// 3. Responsibility: Notifications
// EmailSender is another layer for notifications
type EmailSender interface {
	SendWelcomeEmail(u User) error
}

// SMTPEmailSender will implement EmailSender iface
type SMTPEmailSender struct{}

// SendWelcomeEmail actually sends the email
func (sender *SMTPEmailSender) SendWelcomeEmail(u User) error {
	fmt.Printf("Sending welcome email to %s via SMTP...\n", u.Email)
	return nil
}

// 4. Responsibility: Business Logic Orchestration
// UserService ONLY cares about the rules of registering a user.
type UserService struct {
	repo   UserRepository
	mailer EmailSender
}

// NewUserService is a constructor injecting the dependencies
func NewUserService(r UserRepository, m EmailSender) *UserService {
	return &UserService{
		repo:   r,
		mailer: m,
	}
}

/*
We can swap out the PSQLUserRepository struct with one for MySQL or Mongo if needed
Those funcs will have same signature, inside they will be implementing things differently
Same goes for the SMTPEmailSender struct. We can just swap out for another sender
The key logic is that every layer communicates with the internal layer via interfaces
That way everything is modular, can be testable, and swapped out if needed
*/

// RegisterUser saves a new user in db and sends notification
func (s *UserService) RegisterUser(u User) error {
	// Business rule validation
	if u.Email == "" {
		return errors.New("email cannot be empty")
	}

	// Delegate persistence
	if err := s.repo.Save(u); err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	// Delegate notification
	if err := s.mailer.SendWelcomeEmail(u); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func SRPGood() {
	// Wire up the application
	repo := &PSQLUserRepository{}
	mailer := &SMTPEmailSender{}

	userService := NewUserService(repo, mailer)

	newUser := User{Name: "Alice", Email: "alice@example.com"}
	userService.RegisterUser(newUser)
}

/*
Why the "Good" Way is Better
Need to switch from PostgreSQL to MongoDB? You only write a new MongoUserRepository and inject it. The UserService doesn't change at all. (One reason to change).

Need to switch from SMTP to an API like SendGrid? You only change the email sender component. The UserService remains untouched.

Want to unit test UserService? You can easily pass in "mock" versions of UserRepository and EmailSender that just return nil, allowing you to test the validation logic in isolation without spinning up a real database.
*/
