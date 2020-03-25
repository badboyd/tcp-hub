package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

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
	defer s.Stop()

	if err := s.Start(&net.TCPAddr{Port: *port}); err != nil {
		log.Printf("Cannot start server: %s", err.Error())
		return
	}

	// create a channel to catch interupt signal
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// wait for terminal signal
	<-quit
}
