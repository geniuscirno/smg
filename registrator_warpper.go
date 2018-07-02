package smg

import (
	"errors"
	"fmt"
	"path"

	"github.com/geniuscirno/smg/registrator"
	_ "github.com/geniuscirno/smg/registrator/consul"
	_ "github.com/geniuscirno/smg/registrator/etcd"
)

type appRegistratorWarpper struct {
	registrator registrator.Registrator
	app         *application
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

func newAppRegistratorWarpper(app *application) (*appRegistratorWarpper, error) {
	target, err := parseRegistratorTarget(app.cfg.RegistryUrl)
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

	app.opts.registerEndpoint.Name = app.name
	app.opts.registerEndpoint.Version = app.version
	warpper := &appRegistratorWarpper{app: app}

	warpper.registrator, err = rb.Build(target)
	if err != nil {
		return nil, err
	}
	return warpper, nil
}

func (r *appRegistratorWarpper) Register() error {
	return r.registrator.Register(path.Join("registry", r.app.opts.nameSpace, r.app.name, r.app.version), r.app.opts.registerEndpoint)
}

func (r *appRegistratorWarpper) Degister() error {
	return r.registrator.Degister(path.Join("registry", r.app.opts.nameSpace, r.app.name, r.app.version), r.app.opts.registerEndpoint)
}
