package smg

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/geniuscirno/smg/configurator"
)

type CfgWatcher interface {
	// WatchConfig(key string, value )
}

type appConfiguratorWarpper struct {
	configurator configurator.Configurator
	app          *Application
	cfg          interface{}
}

func parseRegistratorTarget(target string) (configurator.Target, bool) {
	spl := strings.SplitN(target, "://", 2)
	if len(spl) < 2 {
		return configurator.Target{}, false
	}
	return configurator.Target{Scheme: spl[0], Endpoint: spl[1]}, true
}

func newAppConfiguratorWarpper(app *Application) (*appConfiguratorWarpper, error) {
	target, ok := parseRegistratorTarget(app.opts.configuratorUrl)
	if !ok {
		return nil, fmt.Errorf("invalid configurator url: %s", app.opts.configuratorUrl)
	}

	cb, ok := configurator.Get(target.Scheme)
	if !ok {
		return nil, fmt.Errorf("invalid scheme: %s", target.Scheme)
	}

	if app.opts.config == nil {
		return nil, errors.New("WithConfigurator: config is nil")
	}

	warpper := &appConfiguratorWarpper{app: app, cfg: app.opts.config}

	var err error
	warpper.configurator, err = cb.Build(target)
	if err != nil {
		return nil, err
	}
	return warpper, nil
}

func (c *appConfiguratorWarpper) Load() error {
	log.Println("configurator:Load")
	return c.configurator.Load(c.cfg)
}

func (c *appConfiguratorWarpper) Watch() {
	log.Println("configurator:Watch")
	for {
		wc, ok := <-c.configurator.Watch()
		if !ok {
			log.Println("configurator:Watch watch channel closed!")
		}
	}
}
