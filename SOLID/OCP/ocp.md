# Open-Closed Principle

The Open-Closed Principle (OCP) is the second principle in SOLID.

Software entities (classes, structs, packages, modules, functions) should be **open for extension, but closed for modification**.

This means:

- ✅ You should be able to add new behavior
- ❌ Without changing existing code

## How This Appears in a Production Go Project

A typical project structure might look like this:

```
internal/
│
├── domain/
│   └── payment.go
│
├── service/
│   └── payment_service.go
│
├── payment/
│   ├── creditcard.go
│   ├── paypal.go
│   ├── stripe.go
│   └── bitcoin.go
│
└── cmd/
    └── main.go
```

The service package depends only on the interface:

```go
type PaymentProcessor interface {
    Pay(amount float64) error
}
```

Each concrete implementation lives in its own file/package. Adding a new payment method means adding a new implementation without modifying the service.

## OCP with Dependency Injection

OCP is usually paired with dependency injection.

```go
processor := StripeProcessor{}
service := NewPaymentService(processor)
```

or using a factory:

```go
processor := factory.GetProcessor("stripe")
service := NewPaymentService(processor)
```

The service doesn't know or care which implementation it receives — it only relies on the `PaymentProcessor` interface.

## Where OCP Is Used in Real-World Go Code

You'll see this pattern in many production systems:

- **Storage backends** (`S3Storage`, `GCSStorage`, `AzureBlobStorage`) implementing a common `Storage` interface.
- **Authentication providers** (`JWTAuth`, `OAuthAuth`, `APIKeyAuth`) implementing an `Authenticator` interface.
- **Logging** (`ConsoleLogger`, `FileLogger`, `KafkaLogger`) implementing a `Logger` interface.
- **Search engines** (`ElasticSearch`, `OpenSearch`, `Meilisearch`) implementing a `SearchClient` interface.
- **Message queues** (`Kafka`, `RabbitMQ`, `NATS`, `SQS`) implementing a `Publisher`/`Consumer` interface.
- **Caching** (`Redis`, in-memory cache, `Memcached`) implementing a `Cache` interface.

## Final Flow in Our Payments System Example

What happens internally when the application starts:

```
main()
  → imports package "payment"
    → Go executes every init() in that package
      → Stripe init() → Register("stripe", StripeProcessor{})
      → PayPal init() → Register("paypal", PayPalProcessor{})
```

Registry becomes:

```go
map[string]Processor{
    "stripe": StripeProcessor{},
    "paypal": PayPalProcessor{},
}
```

### Tomorrow Product Asks for Razorpay

Create **one** file — `razorpay.go`:

```go
package payment

import "fmt"

type RazorpayProcessor struct{}

func (RazorpayProcessor) Pay(amount float64) error {
    fmt.Printf("Paid %.2f using Razorpay\n", amount)
    return nil
}

func init() {
    Register("razorpay", RazorpayProcessor{})
}
```

No existing code was touched — pure extension.

## A Small Refinement for Production Code

One improvement is to register **constructors** rather than instances. This avoids sharing the same object if a processor has internal state.

```go
type Factory func() Processor

var registry = map[string]Factory{}

func Register(name string, factory Factory) {
    registry[name] = factory
}

func Get(name string) (Processor, error) {
    factory, ok := registry[name]
    if !ok {
        return nil, fmt.Errorf("unknown processor: %s", name)
    }

    return factory(), nil
}
```

Registration becomes:

```go
func init() {
    Register("stripe", func() Processor {
        return &StripeProcessor{}
    })
}
```

Now every call to `Get("stripe")` returns a fresh processor instance, which is safer if processors hold mutable state.
