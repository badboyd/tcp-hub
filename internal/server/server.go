package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/badboyd/tcp-hub/pkg/id"
	"github.com/badboyd/tcp-hub/pkg/message"
)

type client struct {
	id   uint64
	conn net.Conn
}

// Server handles and stores clients information
type Server struct {
	m        sync.RWMutex
	clients  map[uint64]*client
	close    chan struct{}
	idSeq    id.Seq
	listener net.Listener
	wg       sync.WaitGroup
}

// New creates new server
func New() *Server {
	return &Server{
		close:   make(chan struct{}),
		clients: make(map[uint64]*client),
	}
}

// Start server at laddr
func (s *Server) Start(laddr *net.TCPAddr) error {
	log.Println("Start server at ", laddr.String())

	listener, err := net.ListenTCP(laddr.Network(), laddr)
	if err != nil {
		return err
	}
	s.listener = listener

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Stop accepting conn: %s", err.Error())
				return
			}

			cli := &client{
				id:   s.idSeq.Next(),
				conn: conn,
			}

			s.addClient(cli)
			s.wg.Add(1)

			go func() {
				defer s.wg.Done()
				s.handle(cli)
			}()
		}
	}()

	return nil
}

func (s *Server) addClient(cli *client) {
	s.m.Lock()
	defer s.m.Unlock()

	s.clients[cli.id] = cli
}

func (s *Server) removeClient(cli *client) {
	s.m.Lock()
	defer s.m.Unlock()

	cli.conn.Close()
	delete(s.clients, cli.id)
}

func (s *Server) handle(cli *client) {
	defer s.removeClient(cli)

	r := bufio.NewReader(cli.conn)

ReadLoop:
	for {
		select {
		case <-s.close:
			log.Printf("Stop serving client %d\n", cli.id)
			return
		default:
			cli.conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
			line, err := r.ReadString('\n')
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue ReadLoop
				}
				log.Printf("[%d] ReadString error: %s\n", cli.id, err.Error())
				return
			}

			var msg string

			parts := strings.SplitN(line[:len(line)-1], " ", 2)
			switch parts[0] {
			case message.IdentityType:
				msg = fmt.Sprintf("%s %d\n", message.IdentityType, cli.id)
			case message.ListType:
				clientIDs := []uint64{}
				for _, clientID := range s.ListClientIDs() {
					if clientID != cli.id {
						clientIDs = append(clientIDs, clientID)
					}
				}
				msg = fmt.Sprintf("%s %s\n", message.ListType, id.JoinIDArray(clientIDs, ","))
			case message.RelayType:
				var err error
				var size int
				var receivers string

				if _, err = fmt.Sscanf(parts[1], "%s %d", &receivers, &size); err != nil {
					log.Printf("Message in wrong format: %s\n", err.Error())
					return
				}

				receiverIDs, err := id.ConvertFromStringToArray(receivers)
				if err != nil {
					log.Printf("ReceiverIDs in wrong format: %s\n", err.Error())
					return
				}

				data := make([]byte, size)
				if _, err = io.ReadFull(r, data); err != nil {
					log.Printf("Cannot read full data: %s\n", err.Error())
					return
				}
				go s.relayMessage(cli.id, receiverIDs, data)
			default:
				msg = "Unknown message\n"
			}

			if msg != "" {
				_, err = cli.conn.Write([]byte(msg))
			}
			if err != nil {
				log.Printf("Error write message to %d", cli.id)
				return
			}
		}
	}
}

func (s *Server) relayMessage(senderID uint64, clientIDs []uint64, data []byte) {
	s.m.RLock()
	defer s.m.RUnlock()

	msg := fmt.Sprintf("%s %d %d\n%s", message.RelayType, senderID, len(data), string(data))
	for _, clientID := range clientIDs {
		if clientID == senderID {
			continue
		}
		if _, err := s.clients[clientID].conn.Write([]byte(msg)); err != nil {
			log.Printf("Error send msg to %d: %s\n", clientID, err.Error())
		}
	}
}

// ListClientIDs returns all the connecting clientIDs
func (s *Server) ListClientIDs() []uint64 {
	s.m.RLock()
	defer s.m.RUnlock()

	clientIDs := []uint64{}
	for clientID := range s.clients {
		clientIDs = append(clientIDs, clientID)
	}
	sort.Slice(clientIDs, func(i, j int) bool {
		return clientIDs[i] < clientIDs[j]
	})
	return clientIDs
}

// Stop server
func (s *Server) Stop() error {
	log.Println("Stop the server")
	if s.listener != nil {
		s.listener.Close()
	}
	if s.close != nil {
		close(s.close)
	}

	s.wg.Wait()
	return nil
}
