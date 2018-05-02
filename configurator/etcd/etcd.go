package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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

func (c *etcdConfigurator) Upload(metaKey string, localMeta *configurator.Meta, iter configurator.FileIter) error {
	if localMeta == nil {
		return nil
	}

	ops := make([]etcd.Op, 0)

	// meta data
	b, err := json.Marshal(localMeta)
	if err != nil {
		return err
	}
	ops = append(ops, etcd.OpPut(metaKey, string(b)))

	// cfg files
	for {
		key, value, next, err := iter()
		if err != nil {
			return err
		}

		ops = append(ops, etcd.OpPut(key, string(value)))
		if !next {
			break
		}
	}

	for {
		var (
			resp *etcd.TxnResponse
			err  error
		)

		remoteMetaResp, err := c.c.KV.Get(context.TODO(), metaKey)
		if err != nil {
			return err
		}

		// not any online cfg exists.
		if len(remoteMetaResp.Kvs) == 0 {
			// upload local cfg.
			resp, err = c.c.Txn(context.TODO()).
				If(etcd.Compare(etcd.CreateRevision(metaKey), "=", 0)).
				Then(ops...).
				Commit()
		} else {
			// unmarhsal remote cfg.
			remoteMeta := &configurator.Meta{}
			err = json.Unmarshal(remoteMetaResp.Kvs[0].Value, remoteMeta)
			if err != nil {
				return err
			}

			// check if our cfg is newer than remote.
			if localMeta.Version <= remoteMeta.Version {
				return nil
			}
			resp, err = c.c.Txn(context.TODO()).
				If(etcd.Compare(etcd.ModRevision(metaKey), "=", remoteMetaResp.Kvs[0].ModRevision)).
				Then(ops...).
				Commit()
		}

		if err != nil {
			return err
		}

		if resp.Succeeded {
			break
		}
	}
	return nil
}

func (c *etcdConfigurator) Load(file string, v configurator.Loader) error {
	resp, err := c.c.KV.Get(context.TODO(), file)
	if err != nil {
		return err
	}

	if resp.Count == 0 {
		return fmt.Errorf("configurator load not found: %s", file)
	}
	return v.Load(resp.Kvs[0].Value)
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

func (w *Watcher) Next(v configurator.Loader) error {
	if w.wch == nil {
		resp, err := w.c.Get(w.ctx, w.file)
		if err != nil {
			return err
		}

		if resp.Count != 0 {
			if err := v.Load(resp.Kvs[0].Value); err != nil {
				return err
			}
		}
		w.wch = w.c.Watch(w.ctx, w.file)
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
			err = v.Load(e.Kv.Value)
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
