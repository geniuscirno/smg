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
	Build(Target, Configer) (Configurator, error)
	Scheme() string
}

type Configer interface {
	Load([]byte) error
	OnConfigChange() error
}

type Configurator interface {
	Load() error
	Watch()
}
