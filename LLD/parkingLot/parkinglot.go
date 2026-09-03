package parkinglot

import (
	"errors"
	"sync"
	"time"
)

type ParkingLot struct {
	mu            sync.RWMutex
	floors        []*Floor
	activeTickets map[string]*Ticket
	fee           int64
}

var sizeMap map[VehicleSize]SpotSize = map[VehicleSize]SpotSize{
	MotorCycle: Small,
	Car:        Medium,
	SUV:        Large,
}

func NewParkingLot(fee int64) *ParkingLot {
	return &ParkingLot{
		floors:        []*Floor{},
		activeTickets: make(map[string]*Ticket),
		fee:           fee,
	}
}

func (pl *ParkingLot) Entry(vehicleSize VehicleSize) *Ticket {
	spot := pl.findAvailableSpot(vehicleSize)
	if spot == nil {
		return nil
	}

	ticket := NewTicket(spot.GetID(), vehicleSize)

	pl.mu.Lock()
	pl.floors[0].occupiedSpotIDs[spot.GetID()] = true
	pl.activeTickets[ticket.GetID()] = ticket
	pl.mu.Unlock()

	return ticket
}

func (pl *ParkingLot) findAvailableSpot(vehicleSize VehicleSize) *Spot {
	spotSize, exists := sizeMap[vehicleSize]
	if !exists {
		return nil
	}

	pl.mu.RLock()
	defer pl.mu.Unlock()
	for _, floor := range pl.floors {
		spot := floor.findAvailableSpot(spotSize)
		if spot != nil {
			return spot
		}
	}

	return nil
}

func (pl *ParkingLot) Exit(ticket *Ticket) (int64, error) {
	_, exists := pl.activeTickets[ticket.GetID()]
	if !exists {
		return 0, errors.New("ticket not found")
	}

	timeElapsed := time.Since(ticket.GetEntryTime())
	fees := pl.calculateFee(timeElapsed)

	delete(pl.floors[ticket.floorNumber].occupiedSpotIDs, ticket.GetSpotID())
	delete(pl.activeTickets, ticket.GetID())

	return fees, nil
}

func (pl *ParkingLot) calculateFee(timeElapse time.Duration) int64 {
	hours := timeElapse.Milliseconds() / (60 * 60 * 1000)
	if timeElapse.Milliseconds()%(60*60*1000) != 0 {
		hours++
	}

	return hours * pl.fee
}
