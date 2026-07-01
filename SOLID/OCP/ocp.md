# Open-Closed Principle

The Open-Closed Principle (OCP) is the second principle in SOLID.

Software entities (classes, structs, packages, modules, functions) should be
open for extension, but closed for modification.

This means:

✅ You should be able to add new behavior ❌ Without changing existing code

How this appears in a production Go project A typical project structure might
look like this:

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

```
type PaymentProcessor interface {      
    Pay(amount float64) error  
} 
```

Each concrete implementation lives in its own file/package. Adding a new payment
method means adding a new implementation without modifying the service.

### OCP with Dependency Injection

OCP is usually paired with dependency injection.

```
processor := StripeProcessor{}
service := NewPaymentService(processor)
```

or using a factory:

```
processor := factory.GetProcessor("stripe") service :=
NewPaymentService(processor)
```

The service doesn't know or care which implementation it receives—it only relies
on the PaymentProcessor interface.
