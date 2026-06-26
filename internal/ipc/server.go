package ipc

import (
	"net"
	"os"
	"path/filepath"
)

type Server struct {
	ln net.Listener
}

func NewServer(socketPath string) (*Server, error) {
	if err := os.MkdirAll(filepath.Dir(socketPath), 0o700); err != nil {
		return nil, err
	}
	os.Remove(socketPath)
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}
	return &Server{ln: ln}, nil
}

func (s *Server) Accept() (net.Conn, error) {
	return s.ln.Accept()
}

func (s *Server) Close() error {
	return s.ln.Close()
}
