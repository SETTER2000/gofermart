package server

import (
	"net"
	"strings"
	"time"
)

// Option -.
type Option func(*Server)

// Port -.
//func Port(port string) Option {
//	return func(s *Server) {
//		s.server.Addr = net.JoinHostPort("", port)
//	}
//}

// Host -.
func Host(host string) Option {
	return func(s *Server) {
		domain := strings.Split(host, ":")
		s.server.Addr = net.JoinHostPort(domain[0], domain[1])
	}
}

// ReadTimeout -.
func ReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.server.ReadTimeout = timeout
	}
}

// WriteTimeout -.
func WriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.server.WriteTimeout = timeout
	}
}

// ShutdownTimeout -.
func ShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = timeout
	}
}
