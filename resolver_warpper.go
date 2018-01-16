package smg

import (
	"strings"

	_ "github.com/geniuscirno/smg/resolver/etcd"
	"google.golang.org/grpc/resolver"
)

type appResolverWarpper struct {
	resolver resolver.Resolver
}

func parseResolverTarget(target string) (resolver.Target, bool) {
	spl := strings.SplitN(target, "://", 2)
	if len(spl) < 2 {
		return resolver.Target{}, false
	}
	return resolver.Target{Scheme: spl[0], Endpoint: spl[1]}, true
}

func newAppResolverWarpper(app *Application) (*appResolverWarpper, error) {
	//target, ok := parseResolverTarget(app.opts.resolverUrl)
	return nil, nil
}
