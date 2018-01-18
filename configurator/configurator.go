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
	Scheme   string
	Endpoint string
}

type Builder interface {
	Build(target Target) (Configurator, error)
	Scheme() string
}

type KV struct {
	Key   string
	Value string
}

type Config interface {
	Load([]KV) error
	Reset()
}

type Configurator interface {
	Watch()
	Load(Config) error
}
