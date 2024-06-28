package remote

import (
	"github.com/colin1989/battery/actor"
	"grpc.go4.org"
)

// Config is the configuration for the remote
type Config struct {
	Host                     string
	Port                     int
	AdvertisedHost           string
	ServerOptions            []grpc.ServerOption
	CallOptions              []grpc.CallOption
	DialOptions              []grpc.DialOption
	EndpointWriterBatchSize  int
	EndpointWriterQueueSize  int
	EndpointManagerBatchSize int
	EndpointManagerQueueSize int
	Kinds                    map[string]*actor.Props
	MaxRetryCount            int
}
