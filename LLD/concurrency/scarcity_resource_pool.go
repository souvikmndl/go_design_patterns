package concurrency

type Connection struct{}

func createNewConnection() *Connection {
	return &Connection{}
}
func (conn *Connection) Execute(query string) {}

type ConnectionPool struct {
	connections chan *Connection
}

func NewConnectionPool(poolSize int) *ConnectionPool {
	pool := &ConnectionPool{
		connections: make(chan *Connection, poolSize),
	}
	for i := 0; i < poolSize; i++ {
		pool.connections <- createNewConnection()
	}
	return pool
}

func (p *ConnectionPool) Acquire() *Connection {
	// will block when empty, preventing scarcity issues
	return <-p.connections
}

func (p *ConnectionPool) Release(conn *Connection) {
	p.connections <- conn
}

func (p *ConnectionPool) ExecuteQueury(query string) {
	conn := p.Acquire()
	defer p.Release(conn)
	conn.Execute(query)
}
