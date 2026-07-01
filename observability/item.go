package observability

import "fmt"

type Subject interface {
	register(o Observer)
	deRegister(o Observer)
	updateAvailability()
	notifyAll()
}

type Item struct {
	Name      string
	InStock   bool
	Observers []Observer
}

func newItem(name string) *Item {
	return &Item{
		Name:      name,
		InStock:   false,
		Observers: []Observer{},
	}
}

func (i *Item) register(o Observer) {
	fmt.Printf("registering %s for %s\n", o.getID(), i.Name)
	i.Observers = append(i.Observers, o)
}

func (i *Item) deRegister(o Observer) {
	fmt.Printf("deregistering %s for %s\n", o.getID(), i.Name)
	for idx, obs := range i.Observers {
		if obs.getID() == o.getID() {
			i.Observers = append(i.Observers[:idx], i.Observers[idx:]...)
		}
	}
}

func (i *Item) updateAvailability() {
	fmt.Println("Item is available now")
	i.InStock = true
	i.notifyAll()
}

func (i *Item) notifyAll() {
	for _, obs := range i.Observers {
		obs.notify(i.Name)
	}
}
