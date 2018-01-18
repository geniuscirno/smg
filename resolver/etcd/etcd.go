package etcd

import (
	"context"
	"encoding/json"
	"log"

	"github.com/geniuscirno/smg/registrator"

	"google.golang.org/grpc/resolver"

	etcd "github.com/coreos/etcd/clientv3"
)

type builder struct{}

func init() {
	resolver.Register(&builder{})
}

func (*builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	cli, err := etcd.NewFromURL("http://" + target.Authority)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	r := &etcdResolver{c: cli, cc: cc, target: target.Endpoint, ctx: ctx, cancel: cancel}
	go r.watcher()
	return r, nil
}

func (*builder) Scheme() string {
	return "etcd"
}

type etcdResolver struct {
	c      *etcd.Client
	cc     resolver.ClientConn
	target string
	ctx    context.Context
	cancel context.CancelFunc
	wc     <-chan etcd.WatchResponse
	addrs  []resolver.Address
}

func (r *etcdResolver) ResolveNow(opt resolver.ResolveNowOption) {}

func (r *etcdResolver) Close() {
	r.cancel()
}

func (r *etcdResolver) watcher() {
	for {
		select {
		case <-r.ctx.Done():
			return
		default:
		}
		if r.wc == nil {
			log.Println("reolvser:etcd get ", r.target)
			resp, err := r.c.KV.Get(r.ctx, r.target, etcd.WithPrefix())
			if err != nil {
				log.Println("etcdResolver: failed to get address")
				return
			}
			var ep registrator.Endpoint
			for _, kv := range resp.Kvs {
				if err := json.Unmarshal(kv.Value, &ep); err != nil {
					continue
				}
				r.addrs = append(r.addrs, resolver.Address{Addr: ep.Addr})
			}

			r.wc = r.c.Watch(r.ctx, r.target, etcd.WithPrevKV())
		} else {
			wc, ok := <-r.wc
			if !ok {
				log.Println("etcdResolver: etcd watch channel closed!")
			}

			var ep registrator.Endpoint
			for _, e := range wc.Events {
				switch e.Type {
				case etcd.EventTypePut:
					if err := json.Unmarshal(e.Kv.Value, &ep); err != nil {
						continue
					}
					log.Println("reolvser:etcd watch add", &ep)
					r.addrs = append(r.addrs, resolver.Address{Addr: ep.Addr})
				case etcd.EventTypeDelete:
					if err := json.Unmarshal(e.PrevKv.Value, &ep); err != nil {
						continue
					}
					log.Println("reolvser:etcd watch del", &ep)
					for i, v := range r.addrs {
						if v.Addr == ep.Addr {
							r.addrs = append(r.addrs[:i], r.addrs[i+1:]...)
							break
						}
					}
				}
			}
		}
		log.Println("resolver:etcd addrs", r.addrs)
		r.cc.NewAddress(r.addrs)
	}
}
