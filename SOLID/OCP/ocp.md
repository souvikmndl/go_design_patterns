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

### Where OCP is used in real-world Go code
You'll see this pattern in many production systems:

* Storage backends (S3Storage, GCSStorage, AzureBlobStorage) implementing a common Storage interface.
* Authentication providers (JWTAuth, OAuthAuth, APIKeyAuth) implementing an Authenticator interface.
* Logging (ConsoleLogger, FileLogger, KafkaLogger) implementing a Logger interface.
* Search engines (ElasticSearch, OpenSearch, Meilisearch) implementing a SearchClient interface.
* Message queues (Kafka, RabbitMQ, NATS, SQS) implementing a Publisher or Consumer interface.
* Caching (Redis, in-memory cache, Memcached) implementing a Cache interface.

Final Flow in our payments system example:
What happens internally?

When the application starts:

`main()`
↓
Imports package
payment
↓
Go executes every init() in that package
Stripe `init()`
↓
`Register("stripe", StripeProcessor{})`
↓
PayPal `init()`
↓
`Register("paypal", PayPalProcessor{})`
↓
Registry becomes
```
map[string]Processor{
    "stripe": StripeProcessor{},
    "paypal": PayPalProcessor{},
}
```

## Tomorrow Product asks for Razorpay

Create ONE file. razorpay.go
```
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

## A small refinement for production code

One improvement is to register constructors rather than instances. This avoids sharing the same object if a processor has internal state.
```
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
```
func init() {
    Register("stripe", func() Processor {
        return &StripeProcessor{}
    })
}
```
Now every call to Get("stripe") returns a fresh processor instance, which is safer if processors hold mutable state.