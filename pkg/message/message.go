package message

const (
	// ListType stands for list command
	ListType = "list"
	// ListReplyFmt stands for list command reply format
	ListReplyFmt = "list %s\n" // "list 1,2\n"

	// RelayType stands for relay command
	RelayType = "relay"

	// IdentityType stands for identity command
	IdentityType = "identity"
	// IdentityReplyFmt stands for identity command reply format
	IdentityReplyFmt = "identity %d\n" // "identity 1\n"
)
