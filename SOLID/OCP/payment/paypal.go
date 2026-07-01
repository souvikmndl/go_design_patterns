package payment

import "fmt"

type PayPalProcessor struct{}

func (PayPalProcessor) Pay(amount float64) error {
	fmt.Printf("Paid %.2f using PayPal\n", amount)
	return nil
}

func init() {
	Register("paypal", PayPalProcessor{})
}
