package registry

import (
	"context"
	"io"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Registry interface {
	Register(ctx context.Context, si ServiceInstance) error
	UnRegister(ctx context.Context, si ServiceInstance) error

	ListServices(ctx context.Context, name string) ([]ServiceInstance, error)
	Subscribe(name string) (<-chan Event, error)

	io.Closer
}

type ServiceInstance struct {
	Name string
	Addr string
	Tags []string
}

type Event struct {
	E *clientv3.Event
}
