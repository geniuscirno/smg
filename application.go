package smg

import (
	"errors"
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
	registratorUrl string
	serviceDesc    *registrator.ServiceDesc
}

type ApplicationOption func(*applicationOptions)

func WithRegistrator(s string) ApplicationOption {
	return func(o *applicationOptions) {
		o.registratorUrl = s
	}
}

func WithServiceDesc(s *registrator.ServiceDesc) ApplicationOption {
	return func(o *applicationOptions) {
		o.serviceDesc = s
	}
}

type Application struct {
	opts           applicationOptions
	appRegistrator *registrator.Registrator
}

func NewApplication(opts ...ApplicationOption) (app *Application, err error) {
	app = &Application{}

	for _, opt := range opts {
		opt(&app.opts)
	}

	if app.opts.registratorUrl != "" {
		app.appRegistrator, err = registrator.NewRegistrator(app.opts.registratorUrl)
		if err != nil {
			return nil, err
		}
		if app.opts.serviceDesc == nil {
			return nil, errors.New("WithRegistrator: no ServiceDesc")
		}
	}
	return app, nil
}

func (app *Application) Run(server Server) error {
	if app.appRegistrator != nil {
		if err := app.appRegistrator.Register(app.opts.serviceDesc); err != nil {
			return err
		}
		defer app.appRegistrator.Degister(app.opts.serviceDesc.ID)
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
