package registrator

import (
	"fmt"
	"log"
	"strings"
)

var (
	m = make(map[string]Builder)
)

// Register register a backend.
func Register(b Builder) {
	m[b.Scheme()] = b
}

// Get returns the builder registered with the given scheme.
func Get(scheme string) (Builder, bool) {
	b, ok := m[scheme]
	return b, ok
}

// ServiceDesc detail of a service.
type ServiceDesc struct {
	ID   string
	Name string
	Addr string
}

// Target represents a terget for registrator.
// example:
//		etcd://localhost:2379
type Target struct {
	Scheme   string
	Endpoint string
}

// Builder creates a client.
type Builder interface {
	Build(target Target) (Client, error)
	Scheme() string
}

// Client a client implements with a backend.
type Client interface {
	Register(*ServiceDesc) error
	Degister(string) error
}

// registratorOptions config a Registrator.
type RegistratorOptions struct {
}

//Registrator represents a registrator.
type Registrator struct {
	client       Client
	parsedTarget Target
}

func parseTarget(target string) (Target, bool) {
	spl := strings.SplitN(target, "://", 2)
	if len(spl) < 2 {
		return Target{}, false
	}
	return Target{Scheme: spl[0], Endpoint: spl[1]}, true
}

// NewRegistrator parse target for scheme and create a Registrator instance.
func NewRegistrator(target string) (reg *Registrator, err error) {
	reg = &Registrator{}

	if parsedTarget, ok := parseTarget(target); ok {
		reg.parsedTarget = parsedTarget
	} else {
		return nil, fmt.Errorf("invalid target %s", target)
	}

	if builder, ok := Get(reg.parsedTarget.Scheme); ok {
		reg.client, err = builder.Build(reg.parsedTarget)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid target scheme %s", reg.parsedTarget.Scheme)
	}
	return reg, nil
}

// Register register a service to backend.
func (reg *Registrator) Register(s *ServiceDesc) error {
	log.Println("Registrator::Register", s.ID, s.Name, s.Addr)
	return reg.client.Register(s)
}

// Degister degister a service to backend.
func (reg *Registrator) Degister(id string) error {
	log.Println("Registrator::Degister", id)
	return reg.client.Degister(id)
}
