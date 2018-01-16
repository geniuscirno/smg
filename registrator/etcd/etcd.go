package etcd

import (
	"context"
	"encoding/json"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/geniuscirno/smg/registrator"
)

func init() {
	registrator.Register(&builder{})
}

type builder struct{}

func (b *builder) Build(target registrator.Target) (registrator.Registrator, error) {
	cli, err := etcd.NewFromURL("http://" + target.Endpoint)
	if err != nil {
		return nil, err
	}
	return &etcdRegistrator{client: cli}, nil
}

func (b *builder) Scheme() string {
	return "etcd"
}

type etcdRegistrator struct {
	client *etcd.Client
}

func (c *etcdRegistrator) Register(s *registrator.Endpoint) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	_, err = c.client.KV.Put(context.Background(), "02service/"+s.Name+"/"+s.Addr, string(b))
	return err
}

func (c *etcdRegistrator) Degister(s *registrator.Endpoint) error {
	_, err := c.client.KV.Delete(context.Background(), "02service/"+s.Name+"/"+s.Addr)
	return err
}
