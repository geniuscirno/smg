package smg

import (
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/geniuscirno/smg/configurator"

	"github.com/geniuscirno/smg/registrator"
)

type EnvironmentType int

const (
	EEnvironmentTypeInvaild EnvironmentType = iota
	EEnvironmentTypePublic
	EEnvironmentTypeBeta
	EEnvironmentTypeDev
	EEnvironmentTypeMax
)

type Server interface {
	Serve() error
}

type applicationOptions struct {
	registerEndpoint *registrator.Endpoint
}

type ApplicationOption func(*applicationOptions)

func WithRegistrator(ep *registrator.Endpoint) ApplicationOption {
	return func(o *applicationOptions) {
		o.registerEndpoint = ep
	}
}

type Config struct {
	RegistryUrl string            `json:"registry-url"`
	ResolverUrl string            `json:"resolver-url"`
	Environment string            `json:"environment"`
	Mongo       string            `json:"mongo"`
	RedisDB     map[string]string `json:"redis-db"`
	RedisCache  map[string]string `json:"redis-cache"`
}

func (c *Config) Load(b []byte) error {
	return json.Unmarshal(b, c)
}

func (c *Config) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}

type Application struct {
	opts            applicationOptions
	appRegistrator  *appRegistratorWarpper
	appConfigurator *appConfiguratorWarpper
	configuratorUrl string
	cfg             *Config
	name            string
}

func NewApplication(name string, url string, opts ...ApplicationOption) (app *Application, err error) {
	app = &Application{name: name, cfg: &Config{}}

	app.configuratorUrl = url
	app.appConfigurator, err = newAppConfiguratorWarpper(app)
	if err != nil {
		return nil, err
	}

	if err := app.loadApplicationCfg(); err != nil {
		return nil, err
	}

	if app.Environment() == EEnvironmentTypeInvaild {
		return nil, errors.New("invalid environment type")
	}

	for _, opt := range opts {
		opt(&app.opts)
	}

	if app.opts.registerEndpoint != nil {
		app.appRegistrator, err = newAppRegistratorWarpper(app)
		if err != nil {
			return nil, err
		}
	}

	return app, nil
}

func (app *Application) Environment() EnvironmentType {
	switch app.cfg.Environment {
	case "public":
		return EEnvironmentTypePublic
	case "beta":
		return EEnvironmentTypeBeta
	case "dev":
		return EEnvironmentTypeDev
	default:
		return EEnvironmentTypeInvaild
	}
}

func (app *Application) Run(server Server) error {
	if app.appRegistrator != nil {
		if err := app.appRegistrator.Register(); err != nil {
			return err
		}
		defer app.appRegistrator.Degister()
	}

	go server.Serve()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	select {
	case <-ch:
		logrus.Info("signal captured, exit.")
	}
	return nil
}

func (app *Application) loadApplicationCfg() error {
	return app.appConfigurator.configurator.Load("application/default", app.cfg)
}

func (app *Application) Load(file string, v configurator.Loader) error {
	return app.appConfigurator.configurator.Load(app.name+"/"+file, v)
}

//func (app *Application) Put(file string, v interface{}) error {
//	return app.appConfigurator.configurator.Put(app.name+"/"+file, v)
//}

func (app *Application) Watch(file string) (configurator.Watcher, error) {
	return app.appConfigurator.configurator.Watch(app.name + "/" + file)
}

func (app *Application) ApplicationUrl(name string) string {
	return app.cfg.RegistryUrl + "/" + path.Join("01registry", name)
}

func (app *Application) Cfg() *Config {
	return app.cfg
}
