# Single Responsibility Principle

The Single Responsibility Principle (SRP) is the "S" in SOLID. It can feel like an abstract academic concept, but it is deeply practical once you see it in action.

At its core, SRP states that a module, struct, or package should have **one, and only one, reason to change**. If a piece of code is responsible for too many different things, changing one part of the system risks breaking unrelated parts.

## Why It Is Used

- **Maintainability** — If your database schema changes, you only update the database code. Your business logic or notification systems remain untouched.
- **Testability** — Testing a struct that does exactly one thing is easy. You don't have to mock a database connection just to test if an email address validation works.
- **Readability** — Smaller, tightly focused structs and functions are much easier to read and understand than massive "god objects" spanning thousands of lines.
- **Reusability** — A standalone `EmailSender` component can be used by `UserService`, `BillingService`, and `PasswordResetService`. If email logic is baked directly into `UserService`, it cannot be reused.

## When It Is Used

- **During system design** — When deciding what packages and structs to create. If you find yourself naming a struct `UserDatabaseAndEmailManager`, that's a huge red flag that you're violating SRP.
- **During refactoring** — If you notice you're constantly opening a single Go file to make changes for completely unrelated feature requests (e.g. you open `user.go` to fix an SQL query today, and open `user.go` tomorrow to change an email template), it's time to split that code up.
