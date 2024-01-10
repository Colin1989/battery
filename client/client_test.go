package client

import (
	"log/slog"
	"testing"
	"time"

	"github.com/colin1989/battery/helper"
	"github.com/colin1989/battery/mocks"
	"github.com/colin1989/battery/net/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendRequestShouldTimeout(t *testing.T) {
	c := New(slog.LevelInfo, 100*time.Millisecond)

	mockConn := mocks.NewMockConnector()
	c.conn = mockConn
	go c.pendingRequestsReaper()

	route := message.NewRoute("sometest", "route")
	data := []byte{0x02, 0x03, 0x04}

	m := message.Message{
		Type:  message.Request,
		ID:    1,
		Route: route,
		Data:  data,
		Err:   false,
	}

	pkt, err := c.buildPacket(m)
	assert.NoError(t, err)

	mockConn.On("Write", pkt).Return(len(pkt))

	c.IncomingMsgChan = make(chan *message.Message, 10)

	c.nextID = 0
	c.SendRequest(route.String(), data)

	msg := helper.ShouldEventuallyReceive(t, c.IncomingMsgChan, 2*time.Second).(*message.Message)

	assert.Equal(t, true, msg.Err)

	mock.AssertExpectationsForObjects(t, mockConn)
}
