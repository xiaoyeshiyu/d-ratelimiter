package etcd

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func TestRegistry(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://10.10.10.21:2379"},
	})
	assert.NoError(t, err)

	server, err := NewServer("test_service", client, grpc.NewServer(), &http.Server{})
	assert.NoError(t, err)

	err = server.Start(context.TODO(), ":8080")
	assert.NoError(t, err)

	//err = server.Close(context.TODO())
	//assert.NoError(t, err)
}
