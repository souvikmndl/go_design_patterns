package ocp

import (
	"errors"
	"fmt"
)

type PaymentMethod string

const (
	CreditCard PaymentMethod = "credit_card"
)

type PaymentService struct{}

/*
This violates OCP because whenever we need to add a new payment method
we will need to add another "case" here. This modifies the code, not extends it
Open to Extension, Closed to Modification
*/
func (p PaymentService) Pay(method PaymentMethod, amount float64) error {
	switch method {
	case CreditCard:
		fmt.Println("Paid using Credit Card")
	default:
		return errors.New("unsupported payment method")
	}

	return nil
}
