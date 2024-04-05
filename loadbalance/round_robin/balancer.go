package round_robin

import (
	"sync/atomic"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type Builder struct {
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for c := range info.ReadySCs {
		conns = append(conns, c)
	}

	return &Balancer{connections: conns}
}

type Balancer struct {
	index       uint32
	connections []balancer.SubConn
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	index := atomic.LoadUint32(&b.index) % uint32(len(b.connections))
	return balancer.PickResult{
		SubConn: b.connections[index],
		Done: func(info balancer.DoneInfo) {
			atomic.AddUint32(&b.index, 1)
		},
	}, nil
}
