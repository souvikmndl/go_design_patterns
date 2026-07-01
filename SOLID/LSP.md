# Liskov Substitution Principle

The Liskov Substitution Principle (LSP) is one of the most misunderstood SOLID principles.

### Definition:
Objects of a subtype should be replaceable for objects of the base type without changing the correctness of the program.

Or more simply:

** If I replace a parent with any of its children, my code should still work exactly as expected. **

A bad example (Rectangle/Square)

Most tutorials use this.
```
type Rectangle struct {
    width  int
    height int
}

func (r *Rectangle) SetWidth(w int) {
    r.width = w
}

func (r *Rectangle) SetHeight(h int) {
    r.height = h
}

func (r Rectangle) Area() int {
    return r.width * r.height
}
```
Someone thinks:

"A square is a rectangle."
```
type Square struct {
    Rectangle
}

func (s *Square) SetWidth(w int) {
    s.width = w
    s.height = w
}

func (s *Square) SetHeight(h int) {
    s.width = h
    s.height = h
}
```
Now suppose we have
```
func Resize(r *Rectangle) {
    r.SetWidth(10)
    r.SetHeight(20)

    if r.Area() != 200 {
        panic("unexpected")
    }
}
```
Works for Rectangle.

But not for Square.

`SetWidth(10)`
`Width=10 Height=10`

`SetHeight(20)`
`Width=20 Height=20`

Area = 400

The function expected 200.

Square changed the behavior.

LSP broken.

This example is mathematically correct but not very useful in backend engineering.

Let's build something you'd actually write.

Practical Example: Payment Processors

Suppose we're building an e-commerce system.

Every payment processor implements
```
type PaymentProcessor interface {
    Pay(amount float64) error
}
```
Stripe
```
type Stripe struct{}

func (s Stripe) Pay(amount float64) error {
    fmt.Printf("Stripe paid %.2f\n", amount)
    return nil
}
```
Razorpay
```
type Razorpay struct{}

func (r Razorpay) Pay(amount float64) error {
    fmt.Printf("Razorpay paid %.2f\n", amount)
    return nil
}
```
Business logic
```
func Checkout(processor PaymentProcessor) error {
    return processor.Pay(500)
}
```
Usage

`Checkout(Stripe{})`
`Checkout(Razorpay{})`

Output

Stripe paid 500
Razorpay paid 500

Everything is fine.

Every implementation behaves according to the contract.

Now let's violate LSP

Suppose someone writes

type FakePayment struct{}

and
```
func (f FakePayment) Pay(amount float64) error {

    fmt.Println("Pretending payment succeeded")

    return nil
}
```
Looks okay.

Implements interface.

Compiles.

But...

It never contacts any payment gateway.

No money transferred.

Business thinks payment succeeded.

Orders are shipped.

Revenue = ₹0.

The interface contract wasn't merely

"Implement Pay()"

It was

"Actually attempt to collect payment."

FakePayment violated that contract.

Although it satisfies the compiler...

it breaks the expectations of the caller.

This is an LSP violation.

Another practical violation

Suppose
```
type PaymentProcessor interface {
    Pay(amount float64) error
}
```
Existing implementations
```
type Stripe struct{}
type Razorpay struct{}
```
Both support

₹10
₹100
₹500
₹1000

Now someone adds
```
type CryptoProcessor struct{}
func (c CryptoProcessor) Pay(amount float64) error {

    if amount < 1000 {
        return errors.New("minimum amount is 1000")
    }

    fmt.Println("crypto payment")

    return nil
}
```
Business code
```
func Checkout(p PaymentProcessor) error {
    return p.Pay(100)
}
```
Works for

`Stripe`
`Razorpay`

Fails for

`CryptoProcessor`

because it silently introduced a stricter precondition.
The caller expected
Any payment processor
Instead
Only payments >=1000
Behavior changed.

**LSP broken.**

A better design

Don't force processors with different capabilities into the same abstraction.

Instead
```
type PaymentProcessor interface {
    CanProcess(amount float64) bool
    Pay(amount float64) error
}
```
Crypto
```
func (c CryptoProcessor) CanProcess(amount float64) bool {
    return amount >= 1000
}```

Checkout
```
func Checkout(p PaymentProcessor) error {

    if !p.CanProcess(100) {
        return errors.New("processor cannot handle this amount")
    }

    return p.Pay(100)
}
```

Now every implementation follows the contract.
No surprises.
LSP satisfied.

Another backend example (File Storage)

Suppose
```
type Storage interface {
    Save(name string, data []byte) error
}
```
Implementations

`type S3Storage struct{}`
`type LocalStorage struct{}`

Both
`Save()`
actually persist data.

Now someone creates
```
type ReadOnlyStorage struct{}
func (r ReadOnlyStorage) Save(name string, data []byte) error {
    panic("cannot save")
}
```
Business code
```
func Upload(storage Storage) error {
    return storage.Save("image.jpg", bytes)
}
```
Everything worked before.

Now
`panic`
The subtype is not substitutable.
** LSP violation. **

A better design would separate reading and writing:
```
type Reader interface {
    Read(name string) ([]byte, error)
}

type Writer interface {
    Save(name string, data []byte) error
}
```

Now read-only storage only implements
`Reader`
which is exactly what it supports.

#### A good production-quality example

Imagine an authentication service:

```
type TokenValidator interface {
    Validate(token string) (User, error)
}
```

Implementations:

```
type JWTValidator struct{}
type APIKeyValidator struct{}
type OAuthValidator struct{}
```

Every implementation promises:

* validate credentials,
* return a valid user on success,
* return an error on failure,
* never fabricate a user.

Business code:
```
func Authenticate(v TokenValidator, token string) error {
    user, err := v.Validate(token)
    if err != nil {
        return err
    }

    fmt.Println("Welcome", user.Name)
    return nil
}
```
A bad implementation:
```
type DebugValidator struct{}

func (d DebugValidator) Validate(token string) (User, error) {
    return User{Name: "Admin"}, nil
}
```

It satisfies the interface but violates the semantic contract. Any code that substitutes DebugValidator for another TokenValidator will behave incorrectly in production. This is an LSP violation because callers rely on the interface's behavior, not just its method signatures.

How LSP differs from the other SOLID principles
Principle	Focus	Question it asks
SRP	One responsibility	Does this type have a single reason to change?
OCP	Extensibility	Can I add new behavior without modifying existing code?
LSP	Correct substitutability	Can I replace one implementation with another without breaking callers?
ISP	Small interfaces	Am I forcing clients to depend on methods they don't use?
DIP	Dependency direction	Am I depending on abstractions instead of concrete implementations?
The practical rule to remember

When you define an interface in Go, you're defining more than a method signature—you are defining a behavioral contract. A new implementation should not:

* strengthen the method's preconditions (e.g., require stricter inputs),
* weaken its postconditions (e.g., claim success without doing the work),
* throw unexpected panics where callers expect normal error handling,
* or violate the semantic expectations established by the interface.

If callers can swap implementations without changing their code or getting unexpected behavior, you've respected the Liskov Substitution Principle.