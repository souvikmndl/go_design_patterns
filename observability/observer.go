package observability

import "fmt"

type Observer interface {
	getID() string
	setID(string)
	notify(string)
}

type Customer struct {
	id string
}

func (c *Customer) getID() string {
	return c.id
}

func (c *Customer) setID(id string) {
	c.id = id
}

func (c *Customer) notify(name string) {
	fmt.Printf("Sending email to %s, Item %s is in stock now \n", c.id, name)
}
