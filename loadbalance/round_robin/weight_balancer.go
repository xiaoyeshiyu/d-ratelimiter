package round_robin

import (
	"math"
	"strconv"
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type WeightBalancerBuilder struct {
}

func (w *WeightBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*conn, 0, len(info.ReadySCs))
	for c, cs := range info.ReadySCs {
		weightStr, ok := cs.Address.Attributes.Value("weight").(string)
		if !ok || weightStr == "" {
			continue
		}
		weight, err := strconv.Atoi(weightStr)
		if err != nil {
			continue
		}

		conns = append(conns, &conn{
			c:               c,
			weight:          uint32(weight),
			currentWeight:   0,
			efficientWeight: 0,
		})
	}

	return &WeightBalancer{connections: conns}
}

type WeightBalancer struct {
	connections []*conn
	mutex       sync.Mutex
}

func (w *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if len(w.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	var totalWeight uint32
	var res *conn
	for _, connection := range w.connections {
		connection.mu.Lock()
		totalWeight += connection.efficientWeight
		connection.currentWeight += connection.efficientWeight
		if res == nil || res.currentWeight < connection.currentWeight {
			res = connection
		}
		connection.mu.Unlock()
	}

	res.mu.Lock()
	res.currentWeight -= totalWeight
	res.mu.Unlock()

	return balancer.PickResult{
		SubConn: res.c,
		Done: func(info balancer.DoneInfo) {
			res.mu.Lock()
			defer res.mu.Unlock()
			if info.Err != nil && res.efficientWeight == 0 {
				return
			}

			if info.Err == nil && res.efficientWeight == math.MaxUint32 {
				return
			}

			if info.Err != nil {
				res.efficientWeight--
			} else {
				res.efficientWeight++
			}
		},
		Metadata: nil,
	}, nil
}

type conn struct {
	mu              sync.Mutex
	c               balancer.SubConn
	weight          uint32
	currentWeight   uint32
	efficientWeight uint32
}
