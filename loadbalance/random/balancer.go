package random

import (
	"math/rand"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type BalancerBuilder struct {
}

func (b *BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for c := range info.ReadySCs {
		conns = append(conns, c)
	}

	return &Balancer{connections: conns}
}

type Balancer struct {
	connections []balancer.SubConn
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	r := rand.Intn(len(b.connections))
	return balancer.PickResult{
		SubConn: b.connections[r],
	}, nil
}
