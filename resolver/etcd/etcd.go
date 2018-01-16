package etcd

import (
	"context"
	"encoding/json"
	"errors"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/geniuscirno/smg/registrator"
	"github.com/geniuscirno/smg/resolver"
)

type builder struct{}

func init() {
	resolver.Register(&builder{})
}

func (*builder) Build(target resolver.Target) (resolver.Resolver, error) {
	cli, err := etcd.NewFromURL("http://" + target.Endpoint)
	if err != nil {
		return nil, err
	}
	return &etcdResolver{client: cli}, nil
}

func (*builder) Scheme() string {
	return "etcd"
}

type etcdResolver struct {
	client *etcd.Client
}

func (r *etcdResolver) Resolve(target string) (resolver.Watcher, error) {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Watcher{c: r.client, target: target, ctx: ctx, cancel: cancel}
	return w, nil
}

type Watcher struct {
	c         *etcd.Client
	target    string
	WatchChan <-chan etcd.WatchResponse
	ctx       context.Context
	cancel    context.CancelFunc
}

func (w *Watcher) Next() ([]*resolver.Update, error) {
	if w.WatchChan == nil {
		resp, err := w.c.KV.Get(w.ctx, w.target, etcd.WithPrefix())
		if err != nil {
			return nil, err
		}
		updates := make([]*resolver.Update, 0, len(resp.Kvs))
		var ep registrator.Endpoint
		for _, kv := range resp.Kvs {
			if err := json.Unmarshal(kv.Value, &ep); err != nil {
				continue
			}
			updates = append(updates, &resolver.Update{Op: resolver.Add, Name: ep.Name, Addr: ep.Addr})
		}
		w.WatchChan = w.c.Watch(w.ctx, w.target, etcd.WithPrevKV())
		return updates, nil
	}

	wc, ok := <-w.WatchChan
	if !ok {
		return nil, errors.New("resolver watch channel closed")
	}

	updates := make([]*resolver.Update, 0, len(wc.Events))
	var ep registrator.Endpoint
	for _, e := range wc.Events {
		switch e.Type {
		case etcd.EventTypePut:
			if err := json.Unmarshal(e.Kv.Value, &ep); err != nil {
				continue
			}
		case etcd.EventTypeDelete:
			if err := json.Unmarshal(e.PrevKv.Value, &ep); err != nil {
				continue
			}
		}
		updates = append(updates, &resolver.Update{Op: resolver.Delete, Name: ep.Name, Addr: ep.Addr})
	}
	return updates, nil
}

func (w *Watcher) Close() {
	w.cancel()
}
