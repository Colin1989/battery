package cluster

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
