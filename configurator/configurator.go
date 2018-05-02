package configurator

import (
	"encoding/json"
	"errors"
)

var (
	m             = make(map[string]Builder)
	ErrNotFound   = errors.New("configureurator: not found")
	defaultScheme = "localcfg"
)

func Register(b Builder) {
	m[b.Scheme()] = b
}

func Get(scheme string) (Builder, bool) {
	if b, ok := m[scheme]; ok {
		return b, true
	}
	if b, ok := m[defaultScheme]; ok {
		return b, true
	}
	return nil, false
}

type Target struct {
	Scheme    string
	Authority string
	Endpoint  string
}

type Builder interface {
	Build(Target) (Configurator, error)
	Scheme() string
}

type Loader interface {
	Load([]byte) error
}

type FileIter func() (string, []byte, bool, error)

type Configurator interface {
	Load(string, Loader) error
	Watch(string) (Watcher, error)
	Upload(string, *Meta, FileIter) error
	//Put(path string, v interface{}) error
}

type Watcher interface {
	Next(Loader) error
	Close()
}

type Meta struct {
	Version int64 `json:"version"`
}

func (c *Meta) Load(b []byte) error {
	return json.Unmarshal(b, c)
}

func (c *Meta) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}
