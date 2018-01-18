package smg

import (
	"errors"
	"fmt"

	"github.com/geniuscirno/smg/configurator"
)

type appConfiguratorWarpper struct {
	configurator configurator.Configurator
	app          *Application
}

func parseConfiguratorTarget(target string) (ret configurator.Target, err error) {
	var ok bool
	ret.Scheme, ret.Endpoint, ok = split2(target, "://")
	if !ok {
		return ret, errors.New("parseConfiguratorTarget: invalid target")
	}
	ret.Authority, ret.Endpoint, _ = split2(ret.Endpoint, "/")
	return ret, nil
}

func newAppConfiguratorWarpper(app *Application) (*appConfiguratorWarpper, error) {
	target, err := parseConfiguratorTarget(app.opts.configuratorUrl)
	if err != nil {
		return nil, err
	}

	cb, ok := configurator.Get(target.Scheme)
	if !ok {
		return nil, fmt.Errorf("invalid scheme: %s", target.Scheme)
	}

	if app.opts.config == nil {
		return nil, errors.New("WithConfigurator: config is nil")
	}

	warpper := &appConfiguratorWarpper{app: app}

	warpper.configurator, err = cb.Build(target, app.opts.config)
	if err != nil {
		return nil, err
	}
	if err = warpper.configurator.Load(); err != nil {
		return nil, err
	}
	go warpper.configurator.Watch()

	return warpper, nil
}
