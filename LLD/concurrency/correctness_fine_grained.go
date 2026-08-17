package concurrency

import "sync"

/*
- When coarse grained locking gets in the way of throughput,
    we need fine grained locking.
- We basically have individual locks for each resource/seat/ticket
- And we only lock the exact resource we need
*/

type TicketBookingFineGrained struct {
	locksMu    sync.Mutex
	seatLocks  map[string]*sync.Mutex // 1 lock per seat
	seatOwners sync.Map
}

func NewTicketBookingFineGrained() *TicketBookingFineGrained {
	return &TicketBookingFineGrained{
		seatLocks: make(map[string]*sync.Mutex),
	}
}

func (tb *TicketBookingFineGrained) getLock(seatID string) *sync.Mutex {
	// lock the individual seatlock when getting its lock, other threads
	// might also be looking to lock it
	tb.locksMu.Lock()
	defer tb.locksMu.Unlock()

	if _, exists := tb.seatLocks[seatID]; !exists {
		tb.seatLocks[seatID] = &sync.Mutex{}
	}
	return tb.seatLocks[seatID]
}

func (tb *TicketBookingFineGrained) BookSeat(seatID, visitorID string) bool {
	lock := tb.getLock(seatID)
	lock.Lock() // lock seat
	defer lock.Unlock()

	// if exists it means seat is already booked
	if _, exists := tb.seatOwners.Load(seatID); exists {
		return false
	}

	tb.seatOwners.Store(seatID, visitorID)
	return true
}

// https://www.hellointerview.com/learn/low-level-design/concurrency/correctness#challenges-1
// There can be a deadlock when two users try to swap their seats using fine-grained locking
// Imagine if A wants to swap 7A for 12B while B wants to swap 12B for 7A
// You first lock one, then the other, perform swap then return. A locks 7A and waits for 12B
// B locks 12B and wait for 7A, resulting in a Deadlock, neither can proceeding as each is
// holding what the other wants.
// FIX: Always acquire locks in a consistent order. If every thread locks the smaller seat ID first
// then A and B will both try to lock 12B before 7A. One of them will get the lock, other will wait
// Eventually the swap will complete without deadlock

func (tb *TicketBookingFineGrained) SwapSeats(visitor1, seat1, visitor2, seat2 string) bool {
	// Trivial case
	if seat1 == seat2 {
		return true
	}

	// Always acquire locks in a deterministic order.
	first, second := seat1, seat2
	if first > second {
		first, second = second, first
	}

	firstLock := tb.getLock(first)
	secondLock := tb.getLock(second)

	firstLock.Lock()
	defer firstLock.Unlock()

	secondLock.Lock()
	defer secondLock.Unlock()

	// Validate that the seat assignments haven't changed.
	if tb.seats[seat1] != visitor1 {
		return false
	}
	if tb.seats[seat2] != visitor2 {
		return false
	}

	// Perform the swap.
	tb.seats[seat1], tb.seats[seat2] =
		tb.seats[seat2], tb.seats[seat1]

	return true
}
