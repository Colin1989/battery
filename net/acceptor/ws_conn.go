package acceptor

import (
	"io"
	"net"
	"time"

	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/errors"
	"github.com/colin1989/battery/facade"
	"github.com/colin1989/battery/net/codec"
	"github.com/gorilla/websocket"
)

var _ facade.Connector = (*WSConn)(nil)

type WSConn struct {
	conn   *websocket.Conn
	typ    int // message type
	reader io.Reader
}

// NewWSConn return an initialized *WSConn
func NewWSConn(conn *websocket.Conn) *WSConn {
	w := &WSConn{
		conn: conn,
	}
	return w
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with ConnectTimeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (w *WSConn) Read(b []byte) (int, error) {
	if w.reader == nil {
		t, r, err := w.conn.NextReader()
		if err != nil {
			return 0, err
		}
		w.typ = t
		w.reader = r
	}
	n, err := w.reader.Read(b)
	if err != nil && err != io.EOF {
		return n, err
	} else if err == io.EOF {
		_, r, err := w.conn.NextReader()
		if err != nil {
			return 0, err
		}
		w.reader = r
	}

	return n, nil
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with ConnectTimeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (w *WSConn) Write(b []byte) (int, error) {
	if err := w.conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return 0, err
	}

	return len(b), nil
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (w *WSConn) Close() error {
	return w.conn.Close()
}

// LocalAddr returns the local network address.
func (w *WSConn) LocalAddr() net.Addr {
	return w.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (w *WSConn) RemoteAddr() net.Addr {
	return w.conn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (w *WSConn) SetDeadline(t time.Time) error {
	if err := w.SetReadDeadline(t); err != nil {
		return err
	}

	return w.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (w *WSConn) SetReadDeadline(t time.Time) error {
	return w.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (w *WSConn) SetWriteDeadline(t time.Time) error {
	return w.conn.SetWriteDeadline(t)
}

// GetNextMessage reads the next message available in the stream
func (w *WSConn) GetNextMessage() (b []byte, err error) {
	_, msgBytes, err := w.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	// check head length
	if len(msgBytes) < codec.HeadLength {
		return nil, constant.ErrInvalidPomeloHeader
	}
	header := msgBytes[:codec.HeadLength]
	_, msgSize, err := codec.ParseHeader(header)
	if err != nil {
		return nil, err
	}
	dataLen := len(msgBytes[codec.HeadLength:])
	if dataLen < msgSize {
		return nil, errors.ErrReceivedMsgSmallerThanExpected
	} else if dataLen > msgSize {
		return nil, errors.ErrReceivedMsgBiggerThanExpected
	}
	return msgBytes, err
}
