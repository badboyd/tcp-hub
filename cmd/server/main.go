package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/badboyd/tcp-hub/internal/server"
)

var (
	port = flag.Int("port", 8000, "TCP server port")
)

func init() {
	flag.Parse()
}

func main() {
	s := server.New()

	// handle interupt signal
	go waitForinteruptSignal()

	if err := s.Start(&net.TCPAddr{Port: *port}); err != nil {
		fmt.Printf("Cannot start server: %s", err.Error())
	}
	// wait for terminal signal
}

func waitForinteruptSignal() {

}
