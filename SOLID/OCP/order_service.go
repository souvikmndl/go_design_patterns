package ocp

import "go_design_patterns/solid/ocp/payment"

// OrderService is responsible for payment
type OrderService struct{}

// Checkout func makes the payment using the "method"/processor
func (OrderService) Checkout(method string, amount float64) error {
	// Notice that OrderService doesn't know what Stripe or PayPal are.
	// It only asks the registry for a Processor.
	processor, err := payment.Get(method)
	if err != nil {
		return err
	}

	return processor.Pay(amount)
}
