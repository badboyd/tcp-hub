package client

import (
	"bufio"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const serverPort = 50000

func TestNew(t *testing.T) {
	cli := New()
	defer cli.Close()

	require.NotNil(t, cli)
}

func TestConnect(t *testing.T) {
	serverAddr := net.TCPAddr{Port: serverPort}

	listener, err := net.ListenTCP(serverAddr.Network(), &serverAddr)
	require.NoError(t, err)

	defer listener.Close()

	cli := New()
	defer cli.Close()

	require.NotNil(t, cli)

	require.NoError(t, cli.Connect(&serverAddr))
}

func TestWhoAmI(t *testing.T) {
	cli, srvConn := createTestClient(t)
	defer cli.Close()

	go func() {
		defer srvConn.Close()

		cmd, err := bufio.NewReader(srvConn).ReadString('\n')
		require.NoError(t, err)
		assert.Equal(t, "identity\n", cmd)

		_, err = srvConn.Write([]byte("identity 1\n"))
		require.NoError(t, err)
	}()

	id, err := cli.WhoAmI()
	require.NoError(t, err)
	assert.Equal(t, uint64(1), id)
}

func TestListClientIDs(t *testing.T) {
	cli, srvConn := createTestClient(t)
	defer cli.Close()

	go func() {
		defer srvConn.Close()

		cmd, err := bufio.NewReader(srvConn).ReadString('\n')
		require.NoError(t, err)
		assert.Equal(t, "list\n", cmd)

		_, err = srvConn.Write([]byte("list 1,2\n"))
		require.NoError(t, err)
	}()

	id, err := cli.ListClientIDs()
	require.NoError(t, err)
	assert.Equal(t, []uint64{1, 2}, id)
}

func TestSendMsg(t *testing.T) {
	cli, srvConn := createTestClient(t)
	defer cli.Close()

	go func() {
		defer srvConn.Close()

		expectedMsg := "relay 1,2 5\nhello"
		msg := make([]byte, len(expectedMsg))

		_, err := io.ReadFull(srvConn, msg)
		require.NoError(t, err)
		assert.Equal(t, expectedMsg, string(msg))
	}()

	require.NoError(t, cli.SendMsg([]uint64{1, 2}, []byte("hello")))
}

func TestHandleIncomingMessages(t *testing.T) {
	cli, srvConn := createTestClient(t)
	defer cli.Close()

	go func() {
		defer srvConn.Close()

		_, err := srvConn.Write([]byte("relay 2 5\nhello"))
		require.NoError(t, err)
	}()

	clientChan := make(chan IncomingMessage)
	defer close(clientChan)

	go cli.HandleIncomingMessages(clientChan)

	receivedMsg := <-clientChan
	assert.Equal(t, []byte("hello"), receivedMsg.Body)
	assert.Equal(t, uint64(2), receivedMsg.SenderID)
}

func createTestClient(t *testing.T) (*Client, net.Conn) {
	srvConn, cliConn := net.Pipe()

	cli := Client{
		conn: cliConn,
		r:    bufio.NewReader(cliConn),
	}

	return &cli, srvConn
}
