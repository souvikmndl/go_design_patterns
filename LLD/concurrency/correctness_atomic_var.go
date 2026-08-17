package concurrency

import "sync/atomic"

/*
- Atomic variables are used when you need to READ and MODIFY a single var
- Basically when in a concurrent seat booking system, you want to count
  the number of seats booked. If you do seatsBooked++, it is actually 3 steps:
    - read the count
    - add
    - write to the variable
- Any concurrent thread can interleave and change the data in between
- You need a mechanism to do this Atomically, in 1 step
- We have Atomic Integers, etc for this. It is am instruction on the CPU
    to execute increment this atomically
*/

type BookingStats struct {
	bookedCount int64
}

func (bs *BookingStats) OnSeatBooked() {
	// pass the counter variable as a pointer
	// the type must match EXACTLY
	atomic.AddInt64(&bs.bookedCount, 1)
}

func (bs *BookingStats) GetBookedCount() int64 {
	return atomic.LoadInt64(&bs.bookedCount)
}

// Atomic Compare and Swap (CAS)
/*
- Lets say you want to track the max number of concurrent bookings ever seen
- You can't just set the value as another thread might have already set it higher
- So you read the current value, compute what you want to set, and attempt CAS
- If it fails, becuase another thread already changed it, loop and try again
*/

type ConcurrencyTracker struct {
	maxConcurrent int64
}

func (ct *ConcurrencyTracker) UpdateMaxConcurrent(current int64) {
	// for loop to keep retrying until you return through one of the conditions
	for {
		observed := atomic.LoadInt64(&ct.maxConcurrent)
		if current <= observed {
			//another thread has already updated to higher/equal so no need to update
			return
		}

		if atomic.CompareAndSwapInt64(&ct.maxConcurrent, observed, current) {
			return
		}
		// if false it means CAS failed, try again
	}
}
