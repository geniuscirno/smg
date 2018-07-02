package consul

import (
	"encoding/json"
	"path"

	"github.com/geniuscirno/smg/registrator"
	"github.com/hashicorp/consul/api"
)

func init() {
	registrator.Register(&builder{})
}

type builder struct{}

func (b *builder) Build(target registrator.Target) (registrator.Registrator, error) {
	cli, err := api.NewClient(&api.Config{
		Address: target.Authority,
	})
	if err != nil {
		return nil, err
	}
	return &consulRegistrator{client: cli, target: target}, nil
}

func (b *builder) Scheme() string {
	return "consul"
}

type consulRegistrator struct {
	client *api.Client
	target registrator.Target
}

func (c *consulRegistrator) Register(name string, s *registrator.Endpoint) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	_, err = c.client.KV().Put(&api.KVPair{
		Key:   path.Join(c.target.Endpoint, name, s.Addr),
		Value: b,
	}, nil)
	return err
}

func (c *consulRegistrator) Degister(name string, s *registrator.Endpoint) error {
	_, err := c.client.KV().Delete(path.Join(c.target.Endpoint, name, s.Addr), nil)
	return err
}
