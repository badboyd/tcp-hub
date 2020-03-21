package message

const (
	ListType     = "list"
	RelayType    = "relay"
	IdentityType = "identity"
)

type Message struct {
	Cmd       string
	Sender    uint64
	Body      string
	Receivers []uint64
}

func NewIdentityMessage() *Message {
	return &Message{Cmd: IdentityType}
}

func NewListMessage(sender uint64) *Message {
	return &Message{Cmd: ListType, Sender: sender}
}

func NewRelayMessage(sender uint64, receivers []uint64, body string) *Message {
	return &Message{Cmd: RelayType, Receivers: receivers, Sender: sender}
}
