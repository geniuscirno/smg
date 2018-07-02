package smg

import (
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"path"
	"path/filepath"
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
	registerEndpoint  *registrator.Endpoint
	includeCfgPattern []string
	ignoreCfgPattern  []string
	prefix            string
	cfgPath           string
	nameSpace         string
}

type ApplicationOption func(*applicationOptions)

func WithRegistrator(nameSpace string, advertiseAddr string, meta interface{}) ApplicationOption {
	return func(o *applicationOptions) {
		o.registerEndpoint = &registrator.Endpoint{
			Addr: advertiseAddr,
			Meta: meta,
		}
		o.nameSpace = nameSpace
	}
}

func WithIncludeCfgFile(pattern []string) ApplicationOption {
	return func(o *applicationOptions) {
		o.includeCfgPattern = pattern
	}
}

func WithIgnoreCfgFile(pattern []string) ApplicationOption {
	return func(o *applicationOptions) {
		o.ignoreCfgPattern = pattern
	}
}

func WithPrefix(prefix string) ApplicationOption {
	return func(o *applicationOptions) {
		o.prefix = prefix
	}
}

func WithCfgPath(cfgPath string) ApplicationOption {
	return func(o *applicationOptions) {
		o.cfgPath = cfgPath
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

type Application interface {
	Environment() EnvironmentType
	Run(Server) error
	Load(string, configurator.Loader) error
	Watch(string) (configurator.Watcher, error)
	ApplicationUrl(string, string) string
	Cfg() *Config
}

type application struct {
	opts            applicationOptions
	appRegistrator  *appRegistratorWarpper
	appConfigurator *appConfiguratorWarpper
	parsedTarget    configurator.Target
	cfg             *Config
	name            string
	version         string
}

func NewApplication(name string, version string, url string, opts ...ApplicationOption) (app *application, err error) {
	app = &application{name: name, cfg: &Config{}, version: version}

	for _, opt := range opts {
		opt(&app.opts)
	}

	if app.opts.prefix == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		app.opts.prefix = wd
	}

	if len(app.opts.includeCfgPattern) == 0 {
		app.opts.includeCfgPattern = []string{"*.*", "*"}
	}

	parsedTarget, err := parseConfiguratorTarget(url)
	if err != nil {
		return nil, err
	}

	app.parsedTarget = parsedTarget
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

	if app.opts.registerEndpoint != nil {
		app.appRegistrator, err = newAppRegistratorWarpper(app)
		if err != nil {
			return nil, err
		}
	}

	return app, nil
}

func (app *application) Environment() EnvironmentType {
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

func (app *application) Run(server Server) error {
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

func (app *application) GlobalCfgRoot() string {
	return filepath.ToSlash(path.Join(app.parsedTarget.Endpoint, "cfg"))
}

func (app *application) CfgRoot() string {
	if app.opts.cfgPath == "" {
		return filepath.ToSlash(path.Join(app.GlobalCfgRoot(), app.name, app.version))
	}
	return filepath.ToSlash(path.Join(app.GlobalCfgRoot(), app.opts.cfgPath, app.version))
}

func (app *application) loadApplicationCfg() error {
	return app.appConfigurator.configurator.Load(path.Join(app.GlobalCfgRoot(), "application", "default"), app.cfg)
}

func (app *application) Load(file string, v configurator.Loader) error {
	return app.appConfigurator.configurator.Load(path.Join(app.CfgRoot(), file), v)
}

func (app *application) Watch(file string) (configurator.Watcher, error) {
	return app.appConfigurator.configurator.Watch(path.Join(app.CfgRoot(), file))
}

func (app *application) ApplicationUrl(nameSpace string, name string) string {
	return app.cfg.RegistryUrl + "/" + path.Join("registry", nameSpace, name)
}

func (app *application) Cfg() *Config {
	return app.cfg
}
