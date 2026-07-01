package payment

import "fmt"

type StripeProcessor struct{}

func (StripeProcessor) Pay(amount float64) error {
	fmt.Printf("Paid %.2f using Stripe\n", amount)
	return nil
}

/*
Whenever this package is imported,
Go automatically executes init().
*/
func init() {
	Register("stripe", StripeProcessor{})
}
