package agent

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/blog"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/net/message"
	"github.com/colin1989/battery/net/packet"
)

func processPacket(a *Agent, p *packet.Packet) error {
	switch p.Type {
	case packet.Handshake:
		blog.Debug("Received handshake packet")

		// Parse the json sent with the handshake by the client
		handshakeData := &packet.HandshakeData{}
		if err := json.Unmarshal(p.Data, handshakeData); err != nil {
			defer a.Close()
			blog.Error("Failed to unmarshal handshake data", blog.ErrAttr(err))
			if serr := a.send(herd); serr != nil {
				blog.Error("Error sending handshake error response: %s", blog.ErrAttr(err))
				return err
			}

			return fmt.Errorf("invalid handshake data. Id=%s", a.PID())
		}

		//if err := a.GetSession().ValidateHandshake(handshakeData); err != nil {
		//	defer a.Close()
		//	logger.Log.Errorf("Handshake validation failed: %s", logger.ErrAttr(err))
		//	if serr := a.SendHandshakeErrorResponse(); serr != nil {
		//		logger.Log.Errorf("Error sending handshake error response: %s", logger.ErrAttr(err))
		//		return err
		//	}
		//
		//	return fmt.Errorf("handshake validation failed: %w. SessionId=%d", err, a.GetSession().ID())
		//}

		if err := a.send(hrd); err != nil {
			blog.Error("Error sending handshake response: %s", blog.ErrAttr(err))
			return err
		}
		blog.Debug("Session handshake",
			slog.String("pid", a.PID()), slog.String("addr", a.RemoteAddr().String()))

		a.SetHandshakeData(handshakeData)
		a.SetStatus(constant.StatusHandshake)
		err := a.SetSessionData(constant.IPVersionKey, a.IPVersion())
		if err != nil {
			blog.Warn("failed to save ip version on session", blog.ErrAttr(err))
		}

		blog.Debug("Successfully saved handshake data")

	case packet.HandshakeAck:
		a.SetStatus(constant.StatusWorking)
		blog.Debug("Receive handshake ACK",
			slog.String("pid", a.PID()), slog.String("addr", a.RemoteAddr().String()))

	case packet.Data:
		if !a.CheckStatus(constant.StatusWorking) {
			return fmt.Errorf("receive data on socket which is not yet ACK, session will be closed immediately, remote=%s",
				a.RemoteAddr().String())
		}

		msg, err := message.Decode(p.Data)
		if err != nil {
			return err
		}
		processMessage(a, msg)

	case packet.Heartbeat:
		// expected
	}

	a.SetLastAt()
	return nil
}

func processMessage(a *Agent, msg *message.Message) {

	//if msg.Route.SvType == "" {
	//r.SvType = h.server.Type
	//}

	// TODO 判断是否为 remote
	system := a.ctx.ActorSystem()
	pid := system.NewLocalPID(msg.Route.Service)
	system.Root.Send(pid, actor.WrapEnvelopWithSender(msg, a.pid))
}
