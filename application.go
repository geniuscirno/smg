package smg

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/geniuscirno/smg/configurator"

	"github.com/geniuscirno/smg/registrator"
)

type Server interface {
	Serve() error
}

type ConfiguratorWatcher interface {
	WatchConfig()
}

type applicationOptions struct {
	registratorUrl   string
	registerEndpoint *registrator.Endpoint
	configuratorUrl  string
	config           configurator.Configer
}

type ApplicationOption func(*applicationOptions)

func WithRegistrator(s string, ep *registrator.Endpoint) ApplicationOption {
	return func(o *applicationOptions) {
		o.registratorUrl = s
		o.registerEndpoint = ep
	}
}

func WithConfigurator(s string, c configurator.Configer) ApplicationOption {
	return func(o *applicationOptions) {
		o.configuratorUrl = s
		o.config = c
	}
}

type Application struct {
	opts            applicationOptions
	appRegistrator  *appRegistratorWarpper
	appConfigurator *appConfiguratorWarpper
}

func NewApplication(opts ...ApplicationOption) (app *Application, err error) {
	app = &Application{}

	for _, opt := range opts {
		opt(&app.opts)
	}

	if app.opts.configuratorUrl != "" {
		app.appConfigurator, err = newAppConfiguratorWarpper(app)
		if err != nil {
			return nil, err
		}
	}

	if app.opts.registratorUrl != "" {
		app.appRegistrator, err = newAppRegistratorWarpper(app)
		if err != nil {
			return nil, err
		}
	}

	return app, nil
}

func (app *Application) Run(server Server) error {
	if app.appRegistrator != nil {
		if err := app.appRegistrator.Register(); err != nil {
			return err
		}
		defer app.appRegistrator.Degister()
	}

	if app.appConfigurator != nil {
		go func() {
			if _, ok := server.(CfgWatcher); ok {
				app.appConfigurator.Watch()
			}
		}()
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve()
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	select {
	case <-ch:
		log.Println("signal captured, exit.")
		return nil
	case err := <-errCh:
		return err
	}
}
