package etcd

import (
	"context"
	"reflect"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/geniuscirno/smg/configurator"
)

type builder struct{}

func init() {
	configurator.Register(&builder{})
}

func (*builder) Build(target configurator.Target) (configurator.Configurator, error) {
	c, err := etcd.NewFromURL("http://" + target.Endpoint)
	if err != nil {
		return nil, err
	}
	return &etcdConfigurator{c: c}, nil
}

func (*builder) Scheme() string {
	return "etcd"
}

type etcdConfigurator struct {
	c *etcd.Client
}

func (c *etcdConfigurator) Load(cfg configurator.Config) error {
	t := reflect.ValueOf(cfg).Type()
	var kvs []configurator.KV
	for i := 0; i < t.NumField(); i++ {
		path := t.Field(i).Tag.Get("cfgd")
		if path == "" {
			continue
		}
		resp, err := c.c.KV.Get(context.TODO(), path)
		if err != nil {
			return err
		}

		if len(resp.Kvs) == 0 {
			continue
		}

		kvs = append(kvs, configurator.KV{Key: string(resp.Kvs[0].Key), Value: string(resp.Kvs[0].Value)})
	}
	return cfg.Load(kvs)
}

func (c *etcdConfigurator) Watch(cfg configurator.Config) {
	for {

	}
}
