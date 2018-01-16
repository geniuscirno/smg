package smg

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/geniuscirno/smg/registrator"
)

type Server interface {
	Serve() error
}

type applicationOptions struct {
	registratorUrl   string
	resolverUrl      string
	registerEndpoint *registrator.Endpoint
}

type ApplicationOption func(*applicationOptions)

func WithRegistrator(s string, ep *registrator.Endpoint) ApplicationOption {
	return func(o *applicationOptions) {
		o.registratorUrl = s
		o.registerEndpoint = ep
	}
}

func WithResolver(s string) ApplicationOption {
	return func(o *applicationOptions) {
		o.resolverUrl = s
	}
}

type Application struct {
	opts           applicationOptions
	appRegistrator *appRegistratorWarpper
	appResolver    *appResolverWarpper
}

func NewApplication(opts ...ApplicationOption) (app *Application, err error) {
	app = &Application{}

	for _, opt := range opts {
		opt(&app.opts)
	}

	if app.opts.registratorUrl != "" {
		app.appRegistrator, err = newAppRegistratorWarpper(app)
		if err != nil {
			return nil, err
		}
	}

	if app.opts.resolverUrl != "" {
		app.appResolver, err = newAppResolverWarpper(app)
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
