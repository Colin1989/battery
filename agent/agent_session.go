package agent

import (
	"encoding/json"
	"github.com/colin1989/battery/net/packet"
	"sync/atomic"
)

func (a *Agent) updateEncodedData() error {
	var b []byte
	b, err := json.Marshal(a.session.Data)
	if err != nil {
		return err
	}
	a.encodedData = b
	return nil
}

func (a *Agent) SetHandshakeData(data *packet.HandshakeData) {
	a.handshakeData = data
}

func (a *Agent) SetStatus(status int32) {
	atomic.StoreInt32(&a.state, status)
}

func (a *Agent) CheckStatus(status int32) bool {
	//atomic.StoreInt32(&a.state, status)
	return atomic.LoadInt32(&a.state) == status
}

func (a *Agent) SetSessionData(key, value string) error {
	a.session.Data[key] = value
	return a.updateEncodedData()
}
