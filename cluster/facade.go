package cluster

import (
	"context"

	"github.com/golang/protobuf/proto"
)

type RPCClient interface {
	Send(route string, data []byte) error
	Request(ctx context.Context) (proto.Message, error)
}

// SDListener interface
type SDListener interface {
	AddServer(*Server)
	RemoveServer(*Server)
}

// ServiceDiscovery is the interface for a service discovery client
type ServiceDiscovery interface {
	//GetServersByType(serverType string) (map[string]*Server, error)
	//GetServer(id string) (*Server, error)
	//GetServers() []*Server
	//SyncServers(firstSync bool) error
	//AddListener(listener SDListener)

	Init(c *Cluster) error
	Shutdown() error
}
