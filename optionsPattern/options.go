package optionspattern

import "time"

/*
When configuring structs (like servers, clients, or business services), don't hardcode configurations or use massive, rigid configuration structs. Use Functional Options. This lets you or future contributors add new configuration parameters later without breaking backwards compatibility.
*/

type Server struct {
	port    int
	timeout time.Duration
}

type Option func(*Server)

func WithTimeout(t time.Duration) Option {
	return func(s *Server) {
		s.timeout = t
	}
}

func NewServer(port int, opts ...Option) *Server {
	srv := &Server{port: port, timeout: 30 * time.Second}
	for _, opt := range opts {
		opt(srv)
	}

	return srv
}
