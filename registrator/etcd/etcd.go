package etcd

import (
	"context"
	"encoding/json"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/geniuscirno/registrator/backend"
)

func init() {
	backend.Register(&builder{})
}

type builder struct{}

func (b *builder) Build(target backend.Target) (backend.Client, error) {
	cli, err := etcd.NewFromURL("http://" + target.Endpoint)
	if err != nil {
		return nil, err
	}
	return &etcdClient{client: cli}, nil
}

func (b *builder) Scheme() string {
	return "etcd"
}

type etcdClient struct {
	client *etcd.Client
}

func (c *etcdClient) Register(s *backend.ServiceDesc) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	_, err = c.client.KV.Put(context.Background(), "02service/"+s.ID, string(b))
	return err
}

func (c *etcdClient) Degister(id string) error {
	_, err := c.client.KV.Delete(context.Background(), id)
	return err
}
