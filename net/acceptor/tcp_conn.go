package acceptor

import (
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/net/codec"
	"io"
	"net"
)

var _ Connector = (*TCPConn)(nil)

type TCPConn struct {
	net.Conn
	remoteAddr net.Addr
}

func (tc *TCPConn) RemoteAddr() net.Addr {
	return tc.remoteAddr
}

// GetNextMessage reads the next message available in the stream
func (tc *TCPConn) GetNextMessage() (b []byte, err error) {
	header, err := io.ReadAll(io.LimitReader(tc.Conn, codec.HeadLength))
	if err != nil {
		return nil, err
	}
	// if the header has no data, we can consider it as a closed connection
	if len(header) == 0 {
		return nil, constant.ErrConnectionClosed
	}
	_, size, err := codec.ParseHeader(header)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(io.LimitReader(tc.Conn, int64(size)))
	if err != nil {
		return nil, err
	}
	if len(data) < size {
		return nil, constant.ErrReceivedMsgSmallerThanExpected
	}
	return append(header, data...), nil
}
