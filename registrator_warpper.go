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

func split2(s, sep string) (string, string, bool) {
	spl := strings.SplitN(s, sep, 2)
	if len(spl) < 2 {
		return "", "", false
	}
	return spl[0], spl[1], true
}

func parseRegistratorTarget(target string) (ret registrator.Target, err error) {
	var ok bool
	ret.Scheme, ret.Endpoint, ok = split2(target, "://")
	if !ok {
		return ret, errors.New("parseRegistratorTarget: invalid target")
	}
	ret.Authority, ret.Endpoint, _ = split2(ret.Endpoint, "/")
	return ret, nil
}

func newAppRegistratorWarpper(app *Application) (*appRegistratorWarpper, error) {
	target, err := parseRegistratorTarget(app.opts.registratorUrl)
	if err != nil {
		return nil, err
	}

	rb, ok := registrator.Get(target.Scheme)
	if !ok {
		return nil, fmt.Errorf("invalid scheme: %s", target.Scheme)
	}

	if app.opts.registerEndpoint == nil {
		return nil, errors.New("WithRegistrator:register endpoint is nil")
	}

	warpper := &appRegistratorWarpper{app: app}

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
