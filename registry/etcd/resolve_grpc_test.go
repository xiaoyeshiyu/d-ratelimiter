package etcd

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func TestClient(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://10.10.10.21:2379"},
	})
	assert.NoError(t, err)

	etcdRegistry, err := NewRegistry("/micro", client)
	assert.NoError(t, err)

	_, err = grpc.Dial(fmt.Sprintf("registry://%s", "test_service"), grpc.WithInsecure(),
		grpc.WithResolvers(NewGrpcResolverBuilder(etcdRegistry, 10*time.Second)))
	assert.NoError(t, err)
}
