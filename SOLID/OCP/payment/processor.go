package payment

// Processor shows the behaviour of a payment processor
type Processor interface {
	Pay(amount float64) error
}
