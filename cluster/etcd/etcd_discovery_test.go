package etcd

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/colin1989/battery/config"

	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/cluster"
	"github.com/stretchr/testify/assert"
)

func newClusterForTest(name, addr string, discover cluster.ServiceDiscovery) *cluster.Cluster {
	host, _port, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}
	port, _ := strconv.Atoi(_port)

	system := actor.NewActorSystem()
	clusterConf := cluster.Configure(name, host, port)

	return cluster.New(system, clusterConf, discover)
}

func TestEtcd(t *testing.T) {
	etcdConfig := config.NewDefaultEtcdConfig()
	etcdConfig.Endpoints = []string{"192.168.110.200:2379"}
	etcd, err := NewWithConfig(etcdConfig)
	assert.NoError(t, err)

	c := newClusterForTest("test", "127.0.0.1:12346", etcd)

	//defer
	err = etcd.Init(c)
	assert.NoError(t, err)
	defer etcd.Shutdown()

	//err = etcd.deregisterService()
	//assert.NoError(t, err)
	//err = etcd.deregisterService()
	//assert.NoError(t, err)

	time.Sleep(time.Hour)
}
