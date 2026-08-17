# Correctness

The problem arises when two users try to access the same resource at the same time.

**Example:** In a ticket booking system, two concurrent users first check if a seat is available and then try to book it. One grabs it first, and the other overwrites the record. So user A, who booked first, gets overwritten by user B. A thinks they have a booked ticket, but the record reflects that B has the actual booking. This causes double booking — a **correctness** issue.

Multiple threads can corrupt shared state when the two steps (check availability, then book) happen as separate steps. Another thread can sneak in between and invalidate the state.

## Where This Shows Up

- Ticket booking
- E-commerce (check stock, then checkout) / inventory management
- Rate limiters (check if limit is exceeded, then increment)
- Connection pools (check if a connection is free before allocating)
- Parking lot systems

## Solutions

| Approach | Description |
|----------|-------------|
| **Coarse-grained locking** | Protects all related state with one lock |
| **Fine-grained locking** | Allows concurrent access to independent resources while protecting related ones |
| **Atomic variables** | Work for single variables but fail for multi-field invariants |
| **Thread confinement** | Eliminates concurrency entirely for related data |

## Common Interview Patterns

- **Check-then-act** — e.g. checking if a seat is available before booking it
- **Read-modify-write** — e.g. incrementing a counter
