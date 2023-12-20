package constant

const (
	_ int32 = iota
	// StatusStart status
	StatusStart
	// StatusHandshake status
	StatusHandshake
	// StatusWorking status
	StatusWorking
	// StatusClosed status
	StatusClosed
)

// IP constants
const (
	IPVersionKey = "ipversion"
	IPv4         = "ipv4"
	IPv6         = "ipv6"
)
