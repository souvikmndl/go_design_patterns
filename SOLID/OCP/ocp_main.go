package ocp

import (
	"log"

	_ "go_design_patterns/solid/ocp/payment" // triggers init() for all processors
)

func main() {

	service := OrderService{}

	err := service.Checkout("stripe", 500) // payment made with stripe
	//err := service.Checkout("paypal", 500) // payment made with paypal

	if err != nil {
		log.Fatal(err)
	}
}
