package adapter

import "fmt"

/*
What is Adapter pattern
- structural design pattern
- The Adapter Pattern is also known as Wrapper
- Its used so that two unrelated objects can work together using Adapter
- The thing that joins these unrelated objects is called an Adapter
*/

/*
Four participants while implementation:
- Target Interface: This is the interface which will be used by the clients
- Adapter: This is a wrapper which implements the target interface and modifies
    the specific request available from the Adapter class.
- Adaptee: This is the object which is used by the Adapter to reuse the existing
  functionality and modify them for desired use.
- Client: This will interact with the Adapter
*/

/*
WHY:
- When you dont need to change the existing object or interface rather wants to add
    new functionality on top of what is existing
- Use the adapter patterns when 2 incompatible interfaces should work together
*/

type Boat struct{}

func (w *Boat) travelToDestination() {
	fmt.Println("boat is navigating to destination")
}

type Client struct{}

func (c *Client) StartingMyJourney(com Transportation) {
	fmt.Println("Starting navigation process")
	com.Travel()
}

type Transportation interface {
	Travel()
}

type Car struct{}

func (m *Car) Travel() {
	fmt.Println("car is navigating to destination")
}

type BoatAdapter struct {
	boat *Boat
}

func (w *BoatAdapter) Travel() {
	fmt.Println("Adaptor used to move boat on roads")
	w.boat.travelToDestination()
}

func adapterExample() {
	client := &Client{}
	car := &Car{}

	fmt.Println("car started")
	client.StartingMyJourney(car)

	fmt.Println("boat started")
	boat := &Boat{}
	boatAdaptor := &BoatAdapter{
		boat: boat,
	}

	client.StartingMyJourney(boatAdaptor)

}
