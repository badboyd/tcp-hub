package main

import (
	"flag"
	"log"
	"net"

	"github.com/badboyd/tcp-hub/internal/client"
	"github.com/badboyd/tcp-hub/pkg/id"
	"github.com/badboyd/tcp-hub/pkg/message"
)

var (
	ip    = flag.String("ip", "127.0.0.1", "TCP Server IP")
	port  = flag.Int("port", 8000, "TCP server port")
	cmd   = flag.String("cmd", "identity", "Command (identity, list, relay)")
	recvs = flag.String("recvs", "", "List of receivers(uint 64) separated by comma")
	msg   = flag.String("msg", "", "Message for relay cmd")
)

func init() {
	flag.Parse()
}

func main() {
	cli := client.New()
	defer cli.Close()

	serverAddr := net.TCPAddr{IP: net.ParseIP(*ip), Port: *port}
	if err := cli.Connect(&serverAddr); err != nil {
		log.Println("Cannot connet to server: ", err.Error())
	}

	switch *cmd {
	case message.IdentityType:
		clientID, err := cli.WhoAmI()
		if err != nil {
			log.Println("Cannot get identity: ", err.Error())
			return
		}

		log.Println("ClientID is: ", clientID)
	case message.ListType:
		clientIDs, err := cli.ListClientIDs()
		if err != nil {
			log.Println("Cannot get identity: ", err.Error())
			return
		}

		log.Println("Other clientIDs are: ", clientIDs)
	case message.RelayType:
		if *recvs == "" || *msg == "" {
			log.Println("Receivers and Message cannot be empty")
			return
		}
		receiverIDs, err := id.ConvertFromStringToArray(*recvs)
		if err != nil {
			log.Println("Receivers in wrong format: ", err.Error())
			return
		}

		if err = cli.SendMsg(receiverIDs, []byte(*msg)); err != nil {
			log.Println("Receivers in wrong format: ", err.Error())
			return
		}
	default:
		log.Println("Unknown cmd: ", *cmd)
	}
}
