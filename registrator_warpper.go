package smg

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/geniuscirno/smg/registrator"
	_ "github.com/geniuscirno/smg/registrator/etcd"
)

type appRegistratorWarpper struct {
	registrator registrator.Registrator
	app         *Application
}

func parseRegistratorTarget(target string) (registrator.Target, bool) {
	spl := strings.SplitN(target, "://", 2)
	if len(spl) < 2 {
		return registrator.Target{}, false
	}
	return registrator.Target{Scheme: spl[0], Endpoint: spl[1]}, true
}

func newAppRegistratorWarpper(app *Application) (*appRegistratorWarpper, error) {
	target, ok := parseRegistratorTarget(app.opts.registratorUrl)
	if !ok {
		return nil, fmt.Errorf("invalid registrator url: %s", app.opts.registratorUrl)
	}

	rb, ok := registrator.Get(target.Scheme)
	if !ok {
		return nil, fmt.Errorf("invalid scheme: %s", target.Scheme)
	}

	if app.opts.registerEndpoint == nil {
		return nil, errors.New("WithRegistrator:register endpoint is nil")
	}

	warpper := &appRegistratorWarpper{app: app}

	var err error
	warpper.registrator, err = rb.Build(target)
	if err != nil {
		return nil, err
	}
	return warpper, nil
}

func (r *appRegistratorWarpper) Register() error {
	log.Println("registrator:Register", r.app.opts.registerEndpoint)
	return r.registrator.Register(r.app.opts.registerEndpoint)
}

func (r *appRegistratorWarpper) Degister() error {
	log.Println("registrator:Degister", r.app.opts.registerEndpoint)
	return r.registrator.Degister(r.app.opts.registerEndpoint)
}
