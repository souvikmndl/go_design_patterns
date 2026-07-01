package hooks

import "fmt"

/*
Internal hooks are used within a single monolithic binary. They allow different parts of your application (or separate internal packages) to register custom behavior that triggers during a specific lifecycle event (e.g., right before saving a user to a database).
*/

// HookFunc signature
type HookFunc func(payload map[string]string) error

// HookRegistry stores all the hooks funcs
type HookRegistry struct {
	hooks map[string][]HookFunc
}

// NewHookRegistry creates a new HookRegistry and returns its pointer
func NewHookRegistry() *HookRegistry {
	return &HookRegistry{
		hooks: make(map[string][]HookFunc),
	}
}

// Register adds a new function to a specific event lifecycle
func (r *HookRegistry) Register(event string, h HookFunc) {
	r.hooks[event] = append(r.hooks[event], h)
}

// Trigger executes all registered functions for an event
func (r *HookRegistry) Trigger(event string, payload map[string]string) error {
	for _, hook := range r.hooks[event] {
		if err := hook(payload); err != nil {
			return err
		}
	}
	return nil
}

// UserService represents core business logic
type UserService struct {
	Registry *HookRegistry
}

// CreateUser is a core businer logic function
func (s *UserService) CreateUser(username, email string) {
	payload := map[string]string{
		"username": username,
		"email":    email,
	}

	s.Registry.Trigger("before_save", payload)
	fmt.Printf("[Core] Saving user %s to the database...\n", username)
	s.Registry.Trigger("after_save", payload)
}

func HooksMain() {
	registry := NewHookRegistry()
	userService := &UserService{Registry: registry}

	// Extension 1: Register an internal logging hook
	registry.Register("before_save", func(p map[string]string) error {
		fmt.Printf("[Hook - Log] Preparing to create account for %s\n", p["username"])
		return nil
	})

	// Extension 2: Register an internal analytics/welcome email hook
	registry.Register("after_save", func(p map[string]string) error {
		fmt.Printf("[Hook - Email] Sending welcome email to %s\n", p["email"])
		return nil
	})

	// Run the core system—it now executes extensions automatically
	userService.CreateUser("johndoe", "john@example.com")
}
