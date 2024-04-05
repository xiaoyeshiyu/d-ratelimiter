package random

import (
	"math/rand"
	"strconv"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type WeightBalancerBuilder struct {
}

func (w *WeightBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]conn, 0, len(info.ReadySCs))
	var totalWeight uint32
	for c, cs := range info.ReadySCs {
		weightStr, ok := cs.Address.Attributes.Value("weight").(string)
		if !ok || weightStr == "" {
			continue
		}
		weight, err := strconv.Atoi(weightStr)
		if err != nil {
			continue
		}

		conns = append(conns, conn{
			c:      c,
			weight: uint32(weight),
		})
		totalWeight += uint32(weight)
	}

	return &WeightBalancer{
		totalWeight: totalWeight,
		connections: conns,
	}
}

type WeightBalancer struct {
	totalWeight uint32
	connections []conn
}

func (w *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(w.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	r := uint32(rand.Intn(int(w.totalWeight)))
	i := 0
	for ; i < len(w.connections); i++ {
		r -= w.connections[i].weight
		if r < 0 {
			break
		}
	}

	return balancer.PickResult{
		SubConn: w.connections[i].c,
	}, nil
}

type conn struct {
	c      balancer.SubConn
	weight uint32
}
