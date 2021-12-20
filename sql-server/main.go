package main

import (
	"flag"
	"log"
)

const (
	nmHost                   = "H"
	nmPort                   = "P"
)

var (
	host             = flag.String(nmHost, "0.0.0.0", "tidb server host")
	port             = flag.String(nmPort, "4000", "tidb server port")
)

func main(){
	flag.Parse()
	s, err := NewServer()
	if err != nil {
		log.Fatalf("failed to init server: %s", err.Error())
	}

	if err != nil{
		log.Fatalf("failed to listen port:  %s", err.Error())
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil{
			log.Printf("failed to accept client conn: %s\n", err.Error())
		}

		cc := s.newConn(conn)

		go s.onConn(cc)
	}
}