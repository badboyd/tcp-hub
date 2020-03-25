package client

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/badboyd/tcp-hub/pkg/id"
	"github.com/badboyd/tcp-hub/pkg/message"
)

type IncomingMessage struct {
	SenderID uint64
	Body     []byte
}

type Client struct {
	id   uint64
	conn net.Conn
	r    *bufio.Reader
}

func New() *Client {
	return &Client{}
}

func (cli *Client) Connect(serverAddr *net.TCPAddr) error {
	conn, err := net.Dial(serverAddr.Network(), serverAddr.String())
	if err != nil {
		return err
	}
	cli.conn = conn
	cli.r = bufio.NewReader(conn)
	return nil
}

func (cli *Client) Close() error {
	if cli.conn != nil {
		return cli.conn.Close()
	}
	return nil
}

func (cli *Client) WhoAmI() (uint64, error) {
	msg := fmt.Sprintf("%s\n", message.IdentityType)
	if _, err := cli.conn.Write([]byte(msg)); err != nil {
		return 0, err
	}

	line, err := cli.r.ReadString('\n')
	if err != nil {
		return 0, err
	}

	var id uint64
	if _, err := fmt.Sscanf(line, message.IdentityReplyFmt, &id); err != nil {
		return 0, err
	}

	cli.id = id
	return id, nil
}

func (cli *Client) ListClientIDs() ([]uint64, error) {
	msg := fmt.Sprintf("%s\n", message.ListType)
	if _, err := cli.conn.Write([]byte(msg)); err != nil {
		return nil, err
	}

	line, err := cli.r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	var clients string
	if _, err := fmt.Sscanf(line, message.ListReplyFmt, &clients); err != nil {
		return nil, err
	}

	return id.ConvertFromStringToArray(clients)
}

func (cli *Client) SendMsg(recipients []uint64, body []byte) error {
	receivers := id.JoinIDArray(recipients, ",")
	msg := fmt.Sprintf("%s %s %d\n%s", message.RelayType, receivers, len(body), string(body))

	_, err := cli.conn.Write([]byte(msg))
	return err
}

func (cli *Client) HandleIncomingMessages(writeCh chan<- IncomingMessage) {
	for {
		line, err := cli.r.ReadString('\n')
		if err != nil {
			log.Printf("[%d] Client error: %s\n", cli.id, err.Error())
			return
		}

		parts := strings.SplitN(line[:len(line)-1], " ", 2)
		switch parts[0] {
		case message.RelayType:
			var size int
			var sender uint64

			if _, err = fmt.Sscanf(parts[1], "%d %d", &sender, &size); err != nil {
				log.Printf("Message in wrong format: %s\n", err.Error())
				return
			}

			data := make([]byte, size)
			if _, err = io.ReadFull(cli.r, data); err != nil {
				log.Printf("Cannot read full data: %s\n", err.Error())
				return
			}

			writeCh <- IncomingMessage{SenderID: sender, Body: data}
		default:
			log.Println("Unknown message")
		}
	}
}
