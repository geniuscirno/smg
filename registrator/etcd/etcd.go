package etcd

import (
	"context"
	"encoding/json"
	"path"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/geniuscirno/smg/registrator"
)

func init() {
	registrator.Register(&builder{})
}

type builder struct{}

func (b *builder) Build(target registrator.Target) (registrator.Registrator, error) {
	cli, err := etcd.NewFromURL("http://" + target.Authority)
	if err != nil {
		return nil, err
	}
	return &etcdRegistrator{client: cli, target: target}, nil
}

func (b *builder) Scheme() string {
	return "etcd"
}

type etcdRegistrator struct {
	client *etcd.Client
	target registrator.Target
}

func (c *etcdRegistrator) Register(name string, s *registrator.Endpoint) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	_, err = c.client.KV.Put(context.Background(), path.Join(c.target.Endpoint, name, s.Addr), string(b))
	return err
}

func (c *etcdRegistrator) Degister(name string, s *registrator.Endpoint) error {
	_, err := c.client.KV.Delete(context.Background(), path.Join(c.target.Endpoint, name, s.Addr))
	return err
}
