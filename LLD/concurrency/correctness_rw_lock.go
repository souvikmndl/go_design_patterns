package concurrency

import "sync"

/*
- when your system is skewed heavily towards reads, it is better to use a RWMutex
- It will increase your read throughput
- Suppose you have a system that reads 100 times per minute but writes only once
- Or a config that gets updated once a day but read many times
- If we use normal coarse grained lock, it will hold up all threads trying to read
    beacuse a normal Mutex can be held by one thread at a time
- However a Read lock can be held by multiple threads at a time
- The Write lock will be held by one thread at a time
- So anytime a thread tries to read, it will acquire a Read Lock, read it, then unlock
- While a thread holds a Read Lock, that resource can't be updated/written to, but others
    holding a Read lock can read it
- While a Write lock is held over a resource, no one can read from or write to it, ensuring correctness
*/

type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewCache() *Cache {
	return &Cache{data: make(map[string]string)}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()         // read lock
	defer c.mu.RUnlock() // read unlock
	val, ok := c.data[key]
	return val, ok
}

func (c *Cache) Put(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

/*
Read-write locks shine when reads vastly outnumber writes. If you have 1000 read operations per second and 1 write per second, readers almost never block each other. But if reads and writes are roughly equal, the overhead of the fancier lock often makes it slower than a simple mutex.

In interviews, mention read-write locks when the interviewer asks about read-heavy workloads. "If reads dominate and writes are rare, I'd use a read-write lock so readers don't block each other. But if the ratio is close to 50/50, a simple mutex is usually faster."
*/
