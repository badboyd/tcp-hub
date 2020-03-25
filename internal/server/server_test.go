package server

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
	srv := New()
	defer srv.Stop()

	require.NotNil(t, srv)
}

func TestStart(t *testing.T) {
	srv := New()
	defer srv.Stop()

	serverAddr := net.TCPAddr{Port: serverPort}
	require.NoError(t, srv.Start(&serverAddr))
}

func TestListClients(t *testing.T) {
	srv := New()
	defer srv.Stop()

	serverAddr := net.TCPAddr{Port: serverPort}
	require.NoError(t, srv.Start(&serverAddr))

	conn1, err := net.Dial(serverAddr.Network(), serverAddr.String())
	require.NoError(t, err)
	defer conn1.Close()

	conn2, err := net.Dial(serverAddr.Network(), serverAddr.String())
	require.NoError(t, err)
	defer conn2.Close()

	clientIDs := srv.ListClientIDs()
	assert.Equal(t, []uint64{1, 2}, clientIDs)
}

func TestHandle(t *testing.T) {
	tcs := []struct {
		name             string
		msg              string
		expectedReply    string
		hasReply         bool
		expectedRelayMsg string
	}{
		{
			name:          "identity",
			msg:           "identity\n",
			expectedReply: "identity 1\n",
			hasReply:      true,
		},
		{
			name:          "list",
			msg:           "list\n",
			expectedReply: "list 2\n",
			hasReply:      true,
		},
		{
			name:             "relay",
			msg:              "relay 2 5\nhello",
			expectedRelayMsg: "relay 1 5\nhello",
		},
		{
			name:          "unknown cmd",
			msg:           "\n",
			expectedReply: "Unknown message\n",
			hasReply:      true,
		},
	}

	srv := New()
	defer srv.Stop()

	serverAddr := net.TCPAddr{Port: serverPort}
	require.NoError(t, srv.Start(&serverAddr))

	conn1, err := net.Dial(serverAddr.Network(), serverAddr.String())
	require.NoError(t, err)
	defer conn1.Close()

	conn2, err := net.Dial(serverAddr.Network(), serverAddr.String())
	require.NoError(t, err)
	defer conn2.Close()

	for _, tc := range tcs {
		var (
			msg              = tc.msg
			expectedReply    = tc.expectedReply
			hasReply         = tc.hasReply
			expectedRelayMsg = tc.expectedRelayMsg
		)

		t.Run(tc.name, func(t *testing.T) {
			_, err := conn1.Write([]byte(msg))
			require.NoError(t, err)

			if hasReply {
				reply, err := bufio.NewReader(conn1).ReadString('\n')
				require.NoError(t, err)

				assert.Equal(t, expectedReply, reply)
			}

			if expectedRelayMsg != "" {
				relayMsg := make([]byte, len(expectedRelayMsg))

				_, err := io.ReadFull(conn2, relayMsg)
				require.NoError(t, err)

				assert.Equal(t, expectedRelayMsg, string(relayMsg))
			}
		})
	}

}
