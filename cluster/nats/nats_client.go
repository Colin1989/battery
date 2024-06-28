package nats

import (
	"context"

	"github.com/colin1989/battery/protos"

	"github.com/colin1989/battery/errors"

	"github.com/colin1989/battery/cluster"
	"github.com/colin1989/battery/config"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
)

type Client struct {
	conn       *nats.Conn
	NatsConfig config.NatsConfig
}

func NewClient() cluster.RPCClient {
	return &Client{}
}

func (ns *Client) Init() error {
	var err error
	ns.conn, err = nats.Connect(ns.NatsConfig.Address)
	if err != nil {
		return err
	}
	return nil
}

func (ns *Client) Send(route string, data []byte) error {
	return ns.conn.Publish(route, data)
}

func (ns *Client) Request(ctx context.Context, server *cluster.Server) (proto.Message, error) {
	var err error
	parent := ctx
	_ = parent
	//parent, err := tracing.ExtractSpan(ctx)
	//if err != nil {
	//	logger.Log.Warnf("failed to retrieve parent span: %s", err.Error())
	//}
	//tags := opentracing.Tags{
	//	"span.kind":       "client",
	//	"local.id":        ns.server.ID,
	//	"peer.serverType": server.Type,
	//	"peer.id":         server.ID,
	//}
	//ctx = tracing.StartSpan(ctx, "NATS RPC Call", tags, parent)
	//defer tracing.FinishSpan(ctx, err)

	if ns.conn == nil {
		err = errors.ErrRPCClientNotInitialized
		return nil, err
	}

	bytes, err := proto.Marshal(&protos.Session{})
	if err != nil {
		return nil, err
	}

	req, err := ns.conn.Request(getChannel(server.Type, server.ID), bytes, ns.NatsConfig.RequestTimeout)
	if err != nil {
		return nil, err
	}

	res := &protos.Response{}
	err = proto.Unmarshal(req.Data, res)
	if err != nil {
		return nil, err
	}

	if res.Error != nil {
		if res.Error.Code == "" {
			res.Error.Code = errors.ErrUnknownCode
		}
		err = &errors.Error{
			Code:     res.Error.Code,
			Message:  res.Error.Msg,
			Metadata: res.Error.Metadata,
		}
		return nil, err
	}

	return res, nil
}
