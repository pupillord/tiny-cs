package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"strconv"
	"sync"
)

var (
	baseConnID  uint32
	serverPID   int
	osUser      string
)

func init() {
	serverPID = os.Getpid()
	currentUser, err := user.Current()
	if err != nil {
		osUser = ""
	} else {
		osUser = currentUser.Name
	}
}

type Server struct {
	host string
	port uint
	rwlock sync.RWMutex
	clients           map[uint32]*clientConn

	listener net.Listener
}

// NewServer creates a new Server.
func NewServer() (*Server, error) {
	p, err := strconv.Atoi(*port)
	if err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("%s:%d", *host, uint(p))

	listener, err := net.Listen("tcp", addr)

	if err != nil {
		return nil, err
	}

	return &Server{
		host: *host,
		port: uint(p),
		listener: listener,
		clients:           make(map[uint32]*clientConn),
	}, nil
}

// newConn creates a new *clientConn from a net.Conn.
// It allocates a connection ID and random salt data for authentication.
func (s *Server) newConn(conn net.Conn) *clientConn {
	cc := newClientConn(s)
	cc.setConn(conn)
	return cc
}

func (s *Server) onConn(cc *clientConn) {
	ctx := context.WithValue(context.Background(), cc, cc.connectionID)
	if err := cc.handshake(ctx); err != nil {
		log.Printf("failed to handshake with client: %s\n", err.Error())

		return
	}

	s.rwlock.Lock()
	s.clients[cc.connectionID] = cc
	//connections := len(s.clients)
	s.rwlock.Unlock()

	cc.Run(ctx)
}

