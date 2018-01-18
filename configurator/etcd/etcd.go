package etcd

import (
	"context"
	"errors"
	"log"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/geniuscirno/smg/configurator"
)

/*
cfg := &Config{}

`04service/cfg/account`

type Config struct{
	Mongo string	`json:"mongo-addr"`
	Redis string 	`json:"redis-addr"`
	LogLevel uint32 `json:"log-level"`
	mutex sync.RWMutex
}

type (c *Config) Load(b []byte) error{
	mutex.Lock()
	defer mutex.Unlock()

	return json.Unmarshal(b, c)
}

func (c *Config) GetMongo() string{
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.Mongo
}

func (c *Config) GetRedis() string{
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.Redis
}

func (c *Config) GetLogLevel() uint32{
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.LogLevel
}

type Server struct{
	cfg *Config
}

func (s *Server) Load() error{
	return s.cfg.Load()
}

func (s *Server) OnConfigChange() error{
	log.SetLevel(s.cfg.LogLevel())
}


*/

type builder struct{}

func init() {
	configurator.Register(&builder{})
}

func (*builder) Build(target configurator.Target, cfg configurator.Configer) (configurator.Configurator, error) {
	c, err := etcd.NewFromURL("http://" + target.Authority)
	if err != nil {
		return nil, err
	}

	return &etcdConfigurator{c: c, cfg: cfg, target: target.Endpoint}, nil
}

func (*builder) Scheme() string {
	return "etcd"
}

type etcdConfigurator struct {
	c      *etcd.Client
	target string
	cfg    configurator.Configer
}

func (c *etcdConfigurator) Load() error {
	resp, err := c.c.KV.Get(context.TODO(), c.target)
	if err != nil {
		return err
	}

	if resp.Count == 0 {
		return errors.New("configurator:load not found")
	}

	return c.cfg.Load(resp.Kvs[0].Value)
}

func (c *etcdConfigurator) Watch() {
	watcher := c.c.Watch(context.TODO(), c.target)
	for {
		wc, ok := <-watcher
		if !ok {
			log.Println("configurator:watch channel closed!")
			return
		}

		for _, e := range wc.Events {
			switch e.Type {
			case etcd.EventTypePut:
				if err := c.cfg.Load(e.Kv.Value); err != nil {
					log.Println(err)
					continue
				}
				c.cfg.OnConfigChange()
			}
		}
	}
}
