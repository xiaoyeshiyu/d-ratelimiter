package etcd

import (
	"context"
	"net"
	"net/http"
	_ "net/http/pprof"

	"github.com/xiaoyeshiyu/micro-tools/registry"
	"go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Server struct {
	registry   registry.Registry
	name       string
	grpcServer *grpc.Server
	httpServer *http.Server
	listener   net.Listener
}

func NewServer(name string, client *clientv3.Client, grpcServer *grpc.Server, httpServer *http.Server) (*Server, error) {
	etcdRegistry, err := NewRegistry("/micro", client)
	if err != nil {
		return nil, err
	}

	return &Server{
		registry:   etcdRegistry,
		name:       name,
		grpcServer: grpcServer,
		httpServer: httpServer,
	}, nil
}

func (s *Server) Start(ctx context.Context, addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	if s.registry != nil {
		si := registry.ServiceInstance{
			Name: s.name,
			Addr: addr,
		}
		err = s.registry.Register(ctx, si)
		if err != nil {
			return err
		}

		defer func() {
			_ = s.registry.UnRegister(context.Background(), si)
			_ = s.registry.Close()
		}()
	}

	eg, ctx := errgroup.WithContext(ctx)
	if s.grpcServer != nil {
		eg.Go(func() error {
			return s.grpcServer.Serve(listener)
		})
	}
	if s.httpServer != nil {
		eg.Go(func() error {
			return s.httpServer.Serve(listener)
		})
	}

	return eg.Wait()
}

func (s *Server) Close(ctx context.Context) error {
	if s.registry != nil {
		_ = s.registry.Close()
	}

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	if s.httpServer != nil {
		_ = s.httpServer.Shutdown(ctx)
	}

	return s.listener.Close()
}
