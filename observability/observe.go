package observability

func Observe() {
	shirt := newItem("us-polo")

	observer1 := &Customer{id: "abc@gmail.com"}
	observer2 := &Customer{id: "def@gmail.com"}

	shirt.register(observer1)
	shirt.register(observer2)

	shirt.updateAvailability()
}
