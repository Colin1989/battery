package actor

import (
	"log/slog"
	"sync"
	"sync/atomic"

	cmap "github.com/orcaman/concurrent-map"
	murmur32 "github.com/twmb/murmur3"
)

// ProcessRegistry
//
//	@Description: 存储管理actor
type ProcessRegistry struct {
	SequenceID  uint64
	ActorSystem *ActorSystem
	Address     string
	LocalPIDs   *SliceMap
	//RemoteHandlers []AddressResolver
	wg sync.WaitGroup
}

type SliceMap struct {
	LocalPIDs []cmap.ConcurrentMap
}

func newSliceMap() *SliceMap {
	sm := &SliceMap{
		LocalPIDs: make([]cmap.ConcurrentMap, 1024),
	}

	for i := 0; i < len(sm.LocalPIDs); i++ {
		sm.LocalPIDs[i] = cmap.New()
	}

	return sm
}

func (s *SliceMap) GetBucket(key string) cmap.ConcurrentMap {
	hash := murmur32.Sum32([]byte(key))
	index := int(hash) % len(s.LocalPIDs)

	return s.LocalPIDs[index]
}

const (
	localAddress = "nonhost"
)

func NewProcessRegistry(actorSystem *ActorSystem) *ProcessRegistry {
	return &ProcessRegistry{
		ActorSystem: actorSystem,
		Address:     localAddress,
		LocalPIDs:   newSliceMap(),
	}
}

//// An AddressResolver is used to resolve remote actors
//type AddressResolver func(*PID) (Process, bool)
//
//func (pr *ProcessRegistry) RegisterAddressResolver(handler AddressResolver) {
//	pr.RemoteHandlers = append(pr.RemoteHandlers, handler)
//}

const (
	digits = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ~+"
)

func uint64ToId(u uint64) string {
	var buf [13]byte
	i := 13
	// base is power of 2: use shifts and masks instead of / and %
	for u >= 64 {
		i--
		buf[i] = digits[uintptr(u)&0x3f]
		u >>= 6
	}
	// u < base
	i--
	buf[i] = digits[uintptr(u)]
	i--
	buf[i] = '$'

	return string(buf[i:])
}

func (pr *ProcessRegistry) NextId() string {
	counter := atomic.AddUint64(&pr.SequenceID, 1)

	return uint64ToId(counter)
}

func (pr *ProcessRegistry) Add(process Process, id string) (*PID, bool) {
	bucket := pr.LocalPIDs.GetBucket(id)
	pid := &PID{
		Address: pr.Address,
		ID:      id,
	}
	absent := bucket.SetIfAbsent(id, process)

	if absent {
		pr.wg.Add(1)
		pr.ActorSystem.Logger().Debug("Add PID", slog.String("pid", pid.String()))
	}

	return pid, absent
}

func (pr *ProcessRegistry) Remove(pid *PID) {
	bucket := pr.LocalPIDs.GetBucket(pid.ID)

	ref, _ := bucket.Pop(pid.ID)
	if l, ok := ref.(*ActorProcess); ok {
		atomic.StoreInt32(&l.dead, 1)
	}
	pr.wg.Done()
	pr.ActorSystem.Logger().Debug("Remove PID", slog.String("pid", pid.String()))
}

func (pr *ProcessRegistry) Get(pid *PID) (Process, bool) {
	if pid == nil {
		return pr.ActorSystem.DeadLetter, false
	}

	// TODO Get Remote Process

	bucket := pr.LocalPIDs.GetBucket(pid.ID)
	ref, ok := bucket.Get(pid.ID)
	if !ok {
		return pr.ActorSystem.DeadLetter, false
	}
	p, ok := ref.(Process)
	return p, ok
}

func (pr *ProcessRegistry) shutdown() {
	pr.wg.Wait()
}
