package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/badboyd/tcp-hub/pkg/id"
	"github.com/badboyd/tcp-hub/pkg/message"
)

type client struct {
	id   uint64
	conn net.Conn
}

type Server struct {
	m       sync.RWMutex
	clients map[uint64]*client
	close   chan struct{}
	idSeq   id.Seq
}

func New() *Server {
	return &Server{
		close:   make(chan struct{}),
		clients: make(map[uint64]*client),
	}
}

func (s *Server) newClient(conn net.Conn) *client {
	return &client{
		conn: conn,
		id:   s.idSeq.Next(),
	}
}

func (s *Server) Start(laddr *net.TCPAddr) error {
	log.Println("Start server at ", laddr.String())

	listener, err := net.ListenTCP(laddr.Network(), laddr)
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		client := s.newClient(conn)
		s.addClient(client)

		go s.handle(client) //
	}

	// wait for interupt signal
	return nil
}

func (s *Server) addClient(c *client) {
	s.m.Lock()
	defer s.m.Unlock()

	s.clients[c.id] = c
}

func (s *Server) removeClient(c *client) {
	s.m.Lock()
	defer s.m.Unlock()

	c.conn.Close()
	delete(s.clients, c.id)
}

func (c *client) getMessage() (*message.Message, error) {
	bytes := make([]byte, 1100000) // because we have maximum 255 ids and 1024 KB in the payload
	datalen, err := c.conn.Read(bytes)
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(string(bytes[:datalen]), "\n", 3)
	switch parts[0] {
	case message.IdentityType:
		return message.NewIdentityMessage(), nil
	case message.ListType:
		return message.NewListMessage(c.id), nil
	case message.RelayType:
		if len(parts) < 3 {
			return nil, fmt.Errorf("Wrong format for reply command")
		}
		receivers := []uint64{}
		for _, word := range strings.Split(parts[1], ",") {
			id, err := strconv.ParseUint(strings.TrimSpace(word), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("Unknown ID format")
			}

			receivers = append(receivers, id)
		}
		return message.NewRelayMessage(c.id, receivers, parts[2]), nil
	default:
		return nil, fmt.Errorf("Unknown message format")
	}

}

// pro
func (s *Server) handle(c *client) {
	defer s.removeClient(c)

	for {
		msg, err := c.getMessage()
		if err != nil {
			// log.Printf("Error get message: %s", err.Error())
			if _, ok := err.(*net.OpError); ok {
				// look like it is a network error
				return
			}
			c.conn.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
			continue
		}

		switch msg.Cmd {
		case message.IdentityType:
			_, err = c.conn.Write([]byte(fmt.Sprintf("%d\n", c.id)))
		case message.ListType:
			// _, err = c.conn.Write([]byte("Implementing"))
			clientIDs := []string{}
			for clientID := range s.clients {
				if clientID != c.id {
					clientIDs = append(clientIDs, fmt.Sprint(clientID))
				}
			}
			_, err = c.conn.Write([]byte(strings.Join(clientIDs, ",")))
		case message.RelayType:
			c.conn.Write([]byte("Implementing"))
			for _, clientID := range msg.Receivers {
				if client := s.clients[clientID]; client != nil {
					_, err = client.conn.Write([]byte(msg.Body))
				}
			}
		}

		if err != nil {
			// log.Printf("Error get message: %s", err.Error())
			if _, ok := err.(*net.OpError); ok {
				// look like it is a network error
				return
			}
			log.Printf("Error %s", err.Error())
		}
	}
}

func (s *Server) ListClientIDs() []uint64 {
	// TODO: Return the IDs of the connected clients
	s.m.RLock()
	defer s.m.RUnlock()

	ids := []uint64{}
	for id := range s.clients {
		ids = append(ids, id)
	}
	return ids
}

func (s *Server) Stop() error {
	// TODO: Stop accepting connections and close the existing ones
	// Graceful shutdown
	close(s.close)
	return nil
}
