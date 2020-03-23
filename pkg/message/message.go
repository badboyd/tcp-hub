package message

const (
	ListType     = "list"
	ListReplyFmt = "list %s\n" // "list 1,2\n"

	RelayType     = "relay"
	RelayReplyFmt = "relay %d %d\n" // "relay 1 5\nhello"

	IdentityType     = "identity"
	IdentityReplyFmt = "identity %d\n" // "identity 1\n"
)
