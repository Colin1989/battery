package etcd

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"github.com/colin1989/battery/blog"
	"github.com/colin1989/battery/cluster"
	"github.com/colin1989/battery/config"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/namespace"
	"google.golang.org/grpc"
)

type Discovery struct {
	cluster       *cluster.Cluster
	cli           *clientv3.Client
	leaseID       clientv3.LeaseID
	keepAliveTTL  time.Duration
	retryInterval time.Duration
	self          *cluster.Server
	servers       map[string]*cluster.Server // all, contains self.
	baseKey       string
	clusterName   string
	revision      uint64
	shutdown      bool
	deregistered  bool
	cancelWatch   func()
	clusterError  error
}

func New() (*Discovery, error) {
	return NewWithConfig(config.DefaultEtcdConfig())
}

func NewWithConfig(cfg config.EtcdConfig) (*Discovery, error) {
	client, err := clientv3.New(loadConfig(cfg))
	if err != nil {
		return nil, err
	}

	client.KV = namespace.NewKV(client.KV, cfg.Prefix)
	client.Watcher = namespace.NewWatcher(client.Watcher, cfg.Prefix)
	client.Lease = namespace.NewLease(client.Lease, cfg.Prefix)

	d := &Discovery{
		cli:     client,
		servers: map[string]*cluster.Server{},
		baseKey: cfg.BaseKey,
	}

	return d, nil
}

func loadConfig(cfg config.EtcdConfig) clientv3.Config {
	conf := clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: cfg.DialTimeout,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}
	if cfg.Username != "" && cfg.Password != "" {
		conf.Username = cfg.Username
		conf.Password = cfg.Password
	}

	return conf
}

func (d *Discovery) Init(c *cluster.Cluster) error {
	if err := d.init(c); err != nil {
		return err
	}

	d.clusterName = c.Config.Name
	servers, err := d.fetchServers()
	if err != nil {
		return err
	}
	d.updateServersWithSelf(servers)
	d.startWatching()

	if err = d.registerService(); err != nil {
		return err
	}
	ctx := context.TODO()
	d.startKeepAlive(ctx)

	return nil
}

func (d *Discovery) Shutdown() error {
	d.shutdown = true
	if d.deregistered {
		err := d.deregisterService()
		if err != nil {
			d.cluster.Logger().Error("deregisterMember", blog.ErrAttr(err))
			return err
		}
		d.deregistered = true
	}
	if d.cancelWatch != nil {
		d.cancelWatch()
	}
	return nil
}

func (d *Discovery) init(c *cluster.Cluster) error {
	d.cluster = c
	addr := d.cluster.ActorSystem.Address()
	host, port, err := splitHostPort(addr)
	if err != nil {
		return err
	}

	memberID := d.cluster.ActorSystem.ID
	d.self = cluster.NewServer(memberID, d.clusterName, host, port)
	d.self.SetMeta("id", d.getID())
	return nil
}

func (d *Discovery) registerService() error {
	data, err := d.self.Serialize()
	if err != nil {
		return err
	}
	leaseID := d.getLeaseID()
	if leaseID <= 0 {
		if leaseID, err = d.newLeaseID(); err != nil {
			return err
		}
		d.setLeaseID(leaseID)
	}

	fullKey := d.buildKey(d.clusterName, d.self.ID)
	_, err = d.cli.Put(context.TODO(), fullKey, string(data), clientv3.WithLease(leaseID))
	if err != nil {
		return err
	}
	return nil
}

func (d *Discovery) deregisterService() error {
	fullKey := d.buildKey(d.clusterName, d.self.ID)
	_, err := d.cli.Delete(context.TODO(), fullKey)
	return err
}

func (d *Discovery) fetchServers() ([]*cluster.Server, error) {
	resp, err := d.cli.Get(context.TODO(), d.baseKey, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	var servers []*cluster.Server
	for _, kv := range resp.Kvs {
		if server, err := cluster.NewServerFromBytes(kv.Value); err != nil {
			return nil, err
		} else {
			servers = append(servers, server)
		}
	}
	d.revision = uint64(resp.Header.GetRevision())
	d.cluster.Logger().Debug("fetch servers",
		slog.Uint64("raft term", resp.Header.GetRaftTerm()),
		slog.Int64("revision", resp.Header.GetRevision()))
	return servers, nil
}

func (d *Discovery) getLeaseID() clientv3.LeaseID {
	return (clientv3.LeaseID)(atomic.LoadInt64((*int64)(&d.leaseID)))
}

func (d *Discovery) setLeaseID(leaseId clientv3.LeaseID) {
	atomic.StoreInt64((*int64)(&d.leaseID), int64(leaseId))
}

func (d *Discovery) newLeaseID() (clientv3.LeaseID, error) {
	ttlSec := d.keepAliveTTL.Seconds()
	resp, err := d.cli.Grant(context.TODO(), int64(ttlSec))
	if err != nil {
		return 0, err
	}
	return resp.ID, nil
}

func (d *Discovery) startKeepAlive(ctx context.Context) {
	go func() {
		for !d.shutdown {
			if err := ctx.Err(); err != nil {
				d.cluster.Logger().Info("Keepalive was stopped.", blog.ErrAttr(err))
				return
			}

			if err := d.keepAliveForever(); err != nil {
				d.cluster.Logger().Info("Failure refreshing service TTL. ReTrying...",
					slog.Duration("after", d.retryInterval),
					blog.ErrAttr(err))
			}
			time.Sleep(d.retryInterval)
		}
	}()
}

func (d *Discovery) keepAliveForever() error {
	if d.self == nil {
		return fmt.Errorf("keepalive must be after initialize")
	}
	data, err := d.self.Serialize()
	if err != nil {
		return err
	}

	var leaseId clientv3.LeaseID
	leaseId, err = d.newLeaseID()
	if err != nil {
		return err
	}
	if leaseId <= 0 {
		return fmt.Errorf("grant lease failed. leaseId=%d", leaseId)
	}
	d.setLeaseID(leaseId)

	key := d.getEtcdKey()
	_, err = d.cli.Put(context.TODO(), key, string(data), clientv3.WithLease(leaseId))
	if err != nil {
		return err
	}
	alive, err := d.cli.KeepAlive(context.TODO(), leaseId)
	if err != nil {
		return err
	}
	for resp := range alive {
		if resp == nil {
			return fmt.Errorf("keep alive failed. resp=%s", resp.String())
		}
		if d.shutdown {
			return nil
		}
	}

	return nil
}

func (d *Discovery) startWatching() {
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)
	d.cancelWatch = cancel
	go func() {
		for !d.shutdown {
			if err := d.keepWatching(ctx); err != nil {
				d.cluster.Logger().Error("Failed to keepWatching.", slog.Any("error", err))
				d.clusterError = err
			}
		}
	}()
}

// GetHealthStatus returns an error if the cluster health status has problems
func (d *Discovery) GetHealthStatus() error {
	return d.clusterError
}

func (d *Discovery) keepWatching(ctx context.Context) error {
	clusterKey := d.buildKey(d.clusterName)
	stream := d.cli.Watch(ctx, clusterKey, clientv3.WithPrefix())
	return d._keepWatching(stream)
}

func (d *Discovery) _keepWatching(stream clientv3.WatchChan) error {
	for resp := range stream {
		if err := resp.Err(); err != nil {
			d.cluster.Logger().Error("Failure watching service.")
			return err
		}
		if len(resp.Events) <= 0 {
			d.cluster.Logger().Error("Empty etcd.events.", slog.Int("events", len(resp.Events)))
			continue
		}
		nodesChanges := d.handleWatchResponse(resp)
		d.updateNodesWithChanges(nodesChanges)
		//p.publishClusterTopologyEvent()
	}
	return nil
}

func (d *Discovery) handleWatchResponse(resp clientv3.WatchResponse) map[string]*cluster.Server {
	changes := map[string]*cluster.Server{}
	for _, ev := range resp.Events {
		key := string(ev.Kv.Key)
		serverID, err := getServerID(key, "/")
		if err != nil {
			d.cluster.Logger().Error("Invalid member.", slog.String("key", key))
			continue
		}

		switch ev.Type {
		case clientv3.EventTypePut:
			server, err := cluster.NewServerFromBytes(ev.Kv.Value)
			if err != nil {
				d.cluster.Logger().Error("Invalid member.", slog.String("key", key))
				continue
			}
			if d.self.Equal(server) {
				d.cluster.Logger().Debug("Skip self.", slog.String("key", key))
				continue
			}
			if _, ok := d.servers[serverID]; ok {
				d.cluster.Logger().Debug("Update member.", slog.String("key", key))
			} else {
				d.cluster.Logger().Debug("New member.", slog.String("key", key))
			}
			changes[serverID] = server
		case clientv3.EventTypeDelete:
			node, ok := d.servers[serverID]
			if !ok {
				continue
			}
			d.cluster.Logger().Debug("Delete member.", slog.String("key", key))
			cloned := *node
			cloned.SetAlive(false)
			changes[serverID] = &cloned
		}
	}
	d.revision = uint64(resp.Header.GetRevision())
	return changes
}

func (d *Discovery) updateServers(servers []*cluster.Server) {
	for _, server := range servers {
		d.servers[server.ID] = server
	}
}

func (d *Discovery) updateServersWithSelf(servers []*cluster.Server) {
	d.updateServers(servers)
	d.servers[d.self.ID] = d.self
}

func (d *Discovery) updateNodesWithChanges(changes map[string]*cluster.Server) {
	for memberId, member := range changes {
		d.servers[memberId] = member
		if !member.IsAlive() {
			delete(d.servers, memberId)
		}
	}
}

func (d *Discovery) getID() string {
	return d.self.ID
}

func (d *Discovery) getEtcdKey() string {
	return d.buildKey(d.clusterName, d.getID())
}

func (d *Discovery) buildKey(keys ...string) string {
	return strings.Join(append([]string{d.baseKey}, keys...), "/")
}
