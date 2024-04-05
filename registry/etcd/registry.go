package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/xiaoyeshiyu/micro-tools/registry"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type Registry struct {
	prefix      string
	client      *clientv3.Client
	session     *concurrency.Session
	mu          sync.Mutex
	watchCancel []context.CancelFunc
}

func NewRegistry(prefix string, client *clientv3.Client) (registry.Registry, error) {
	session, err := concurrency.NewSession(client)
	if err != nil {
		return nil, err
	}

	return &Registry{
		prefix:  prefix,
		client:  client,
		session: session,
	}, nil
}

func (r *Registry) Register(ctx context.Context, si registry.ServiceInstance) error {
	bytes, _ := json.Marshal(si)

	_, err := r.client.Put(ctx, r.instanceKey(si), string(bytes), clientv3.WithLease(r.session.Lease()))
	return err
}

func (r *Registry) UnRegister(ctx context.Context, si registry.ServiceInstance) error {
	_, err := r.client.Delete(ctx, r.instanceKey(si))
	return err
}

func (r *Registry) ListServices(ctx context.Context, name string) ([]registry.ServiceInstance, error) {
	getResponse, err := r.client.Get(ctx, r.serviceKey(name), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	res := make([]registry.ServiceInstance, 0, len(getResponse.Kvs))
	for _, kv := range getResponse.Kvs {
		var si registry.ServiceInstance
		err = json.Unmarshal(kv.Value, &si)
		if err != nil {
			return nil, err
		}
		res = append(res, si)
	}

	return res, nil
}

func (r *Registry) Subscribe(name string) (<-chan registry.Event, error) {
	ctx, cancel := context.WithCancel(context.TODO())

	ctx = clientv3.WithRequireLeader(ctx)
	r.mu.Lock()
	r.watchCancel = append(r.watchCancel, cancel)
	r.mu.Unlock()

	watchChan := r.client.Watch(ctx, r.serviceKey(name), clientv3.WithPrefix())
	res := make(chan registry.Event)
	go func() {
		for {
			select {
			case event := <-watchChan:
				if event.Canceled {
					return
				}

				if event.Err() != nil {
					continue
				}
				for _, e := range event.Events {
					res <- registry.Event{
						E: e,
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return res, nil
}

func (r *Registry) Close() error {
	r.mu.Lock()
	cancels := r.watchCancel
	r.mu.Unlock()
	for _, cancel := range cancels {
		cancel()
	}
	return r.session.Close()
}

func (r *Registry) instanceKey(si registry.ServiceInstance) string {
	return fmt.Sprintf("%s/%s/%s", r.prefix, si.Name, si.Addr)
}

func (r *Registry) serviceKey(name string) string {
	return fmt.Sprintf("%s/%s", r.prefix, name)
}
