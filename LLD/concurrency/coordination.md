# Coordination

This situation arises when a producer needs to hand over time-consuming tasks to workers/consumers. The task is handed over via a queue (in-memory, like a channel, or something like SQS). For LLD purposes we'll consider issues with in-memory channels/queues.

There are three main concerns with this coordination:

## 1. Efficient Waiting

Consumers should sleep when the queue is empty (no work), then wake up immediately when work arrives.

- If they keep polling the queue for messages, they waste CPU cycles doing no work.
- If they sleep for tiny fixed durations (e.g. 100ms) when there is no work, they increase latency — work can arrive the millisecond after the worker goes to sleep.

**Solution:** In Go, goroutines reading from a channel sleep when the channel is empty. Blocking is waiting without consuming resources. As soon as there is a message in the channel, a goroutine wakes up and consumes it automatically.

## 2. Backpressure

Producers should slow down when consumers can't keep up, preventing memory exhaustion.

- If the producer keeps populating the queue faster than consumers can drain it, memory usage keeps growing and can cause errors.

**Solution:** In Go, when a channel is full, it blocks producers from sending more messages until there is space. Use a buffered channel — as soon as space frees up, a producer can send again.

## 3. Thread Safety

The coordination mechanism itself must handle concurrent access without corruption.
