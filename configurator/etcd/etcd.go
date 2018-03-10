package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"path"

	"github.com/sirupsen/logrus"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/geniuscirno/smg/configurator"
)

type builder struct{}

func init() {
	configurator.Register(&builder{})
}

func (*builder) Build(target configurator.Target) (configurator.Configurator, error) {
	c, err := etcd.NewFromURL("http://" + target.Authority)
	if err != nil {
		return nil, err
	}

	return &etcdConfigurator{c: c, target: target}, nil
}

func (*builder) Scheme() string {
	return "etcd"
}

type etcdConfigurator struct {
	c      *etcd.Client
	target configurator.Target
}

func (c *etcdConfigurator) Load(file string, v interface{}) error {
	resp, err := c.c.KV.Get(context.TODO(), path.Join(c.target.Endpoint, "02app/cfg", file))
	if err != nil {
		return err
	}

	if resp.Count == 0 {
		return errors.New("etcdConfigurator:load not found")
	}
	return json.Unmarshal(resp.Kvs[0].Value, v)
}

func (c *etcdConfigurator) Put(file string, v interface{}) error {
	s, err := json.Marshal(v)
	if err != nil {
		return err
	}
	r, err := c.c.KV.Put(context.TODO(), path.Join(c.target.Endpoint, "02app/cfg", file), string(s))
	logrus.Info("KKKKKK", r, err)
	if err != nil {
		return err
	}
	return nil
}

func (c *etcdConfigurator) Watch(file string) (configurator.Watcher, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &Watcher{c: c.c, target: c.target, file: file, ctx: ctx, cancel: cancel}, nil
}

type Watcher struct {
	c      *etcd.Client
	target configurator.Target
	file   string
	ctx    context.Context
	cancel context.CancelFunc
	wch    etcd.WatchChan
}

func (w *Watcher) Next(v interface{}) error {
	if w.wch == nil {
		resp, err := w.c.Get(w.ctx, path.Join(w.target.Endpoint, "02app/cfg", w.file))
		if err != nil {
			return err
		}

		if resp.Count != 0 {
			if err := json.Unmarshal(resp.Kvs[0].Value, v); err != nil {
				return err
			}
		}
		w.wch = w.c.Watch(w.ctx, path.Join(w.target.Endpoint, "02app/cfg", w.file))
		return nil
	}

	wr, ok := <-w.wch
	if !ok {
		return errors.New("configurator watcher close")
	}

	var err error
	for _, e := range wr.Events {
		switch e.Type {
		case etcd.EventTypePut:
			err = json.Unmarshal(e.Kv.Value, v)
			if err != nil {
				continue
			}
		}
	}
	return err
}

func (w *Watcher) Close() {
	w.cancel()
}
