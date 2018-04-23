package configurator

var (
	m = make(map[string]Builder)
)

func Register(b Builder) {
	m[b.Scheme()] = b
}

func Get(scheme string) (Builder, bool) {
	b, ok := m[scheme]
	return b, ok
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

type Configurator interface {
	Load(string, Loader) error
	Watch(string) (Watcher, error)
	//Put(path string, v interface{}) error
}

type Watcher interface {
	Next(Loader) error
	Close()
}
