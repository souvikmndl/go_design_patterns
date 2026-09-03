package parkinglot

import (
	"crypto/rand"
	"time"
)

type VehicleSize string

const (
	MotorCycle VehicleSize = "MOTORCYCLE"
	Car        VehicleSize = "CAR"
	SUV        VehicleSize = "SUV"
)

type Ticket struct {
	id          string
	spotID      string
	floorNumber int
	entryTime   time.Time
	vehicleSize VehicleSize
}

func NewTicket(spotID string, size VehicleSize) *Ticket {
	return &Ticket{
		id:          rand.Text(),
		spotID:      spotID,
		entryTime:   time.Now(),
		vehicleSize: size,
	}
}

func (t Ticket) GetID() string {
	return t.id
}

func (t Ticket) GetSpotID() string {
	return t.spotID
}

func (t Ticket) GetEntryTime() time.Time {
	return t.entryTime
}

func (t Ticket) GetVehicleSize() VehicleSize {
	return t.vehicleSize
}
