package etcd

import (
	"context"
	"time"

	"github.com/xiaoyeshiyu/micro-tools/registry"
	"google.golang.org/grpc/resolver"
)

type GrpcResolverBuilder struct {
	registry registry.Registry
	timeout  time.Duration
}

func NewGrpcResolverBuilder(registry registry.Registry, timeout time.Duration) *GrpcResolverBuilder {
	return &GrpcResolverBuilder{
		registry: registry,
		timeout:  timeout,
	}
}

func (g *GrpcResolverBuilder) Build(
	target resolver.Target,
	cc resolver.ClientConn,
	opts resolver.BuildOptions,
) (resolver.Resolver, error) {
	res := &grpcResolver{
		target:   target,
		client:   cc,
		registry: g.registry,
		timeout:  g.timeout,
	}

	res.resolveNow()
	go res.watch()

	return res, nil
}

func (g *GrpcResolverBuilder) Scheme() string {
	return "registry"
}

type grpcResolver struct {
	target   resolver.Target
	client   resolver.ClientConn
	registry registry.Registry
	close    chan struct{}
	timeout  time.Duration
}

func (g *grpcResolver) ResolveNow(options resolver.ResolveNowOptions) {
	g.resolveNow()
}

func (g *grpcResolver) watch() {
	events, err := g.registry.Subscribe(g.target.Endpoint())
	if err != nil {
		g.client.ReportError(err)
		return
	}

	for {
		select {
		case <-events:
			// todo 可针对事件进行优化
			g.resolveNow()
		case <-g.close:
			return
		}
	}
}

func (g *grpcResolver) resolveNow() {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	services, err := g.registry.ListServices(ctx, g.target.Endpoint())
	if err != nil {
		g.client.ReportError(err)
		return
	}

	addresses := make([]resolver.Address, 0, len(services))
	for _, service := range services {
		addresses = append(addresses, resolver.Address{
			Addr: service.Addr,
		})
	}

	err = g.client.UpdateState(resolver.State{
		Addresses: addresses,
	})
	if err != nil {
		g.client.ReportError(err)
		return
	}
}

func (g *grpcResolver) Close() {
	close(g.close)
}
