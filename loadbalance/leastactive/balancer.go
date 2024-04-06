package leastactive

import (
	"math"
	"sync/atomic"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type BalancerBuilder struct {
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*activeConn, 0, len(info.ReadySCs))
	for c := range info.ReadySCs {
		conns = append(conns, &activeConn{
			c: c,
		})
	}

	return &Balancer{connections: conns}
}

type Balancer struct {
	connections []*activeConn
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	res := &activeConn{count: math.MaxUint64}
	for _, connection := range b.connections {
		count := atomic.LoadUint64(&connection.count)
		if res.count > count {
			res = connection
		}
	}

	atomic.AddUint64(&res.count, uint64(1))
	return balancer.PickResult{
		SubConn: res.c,
		Done: func(info balancer.DoneInfo) {
			atomic.AddUint64(&res.count, uint64(-1))
		},
	}, nil
}

type activeConn struct {
	count uint64
	c     balancer.SubConn
}
