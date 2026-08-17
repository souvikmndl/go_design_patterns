package concurrency

import "sync"

/*
- Solution to the check then act pattern
- Thread will acquire a lock on the resource/state
- All other threads trying to access that resource will have to wait
  until the first thread releases it
- Coarse Grained locking means using one thread to guard ALL booking operations
- This means if thread A is booking 7A, thread B which wants to book 12C will also
  have to wait for the same lock
*/
// Coarse Grained lock is the deault answer to all check then act patterns

// Lock must not released until the entire "check then act" is complete
// One single mutex must be used to lock the entire operation, you can't use
// one lock to check and another to act, that will not create the atomicity of operation
// needed here

// TRADE OFF: Throughput: threads trying to book other seats will have to wait for this one to complete

type TicketBooking struct {
	mu         sync.Mutex
	seatowners map[string]string
}

func NewTicketBooking() *TicketBooking {
	return &TicketBooking{seatowners: make(map[string]string)}
}

func (tb *TicketBooking) BookSeat(seatID, visitorID string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// seat is already booked, return false
	if _, exists := tb.seatowners[seatID]; exists {
		return false
	}
	// book seat
	tb.seatowners[seatID] = visitorID
	return true
}

//
