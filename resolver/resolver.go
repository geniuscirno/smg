package resolver

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

type Operation uint8

const (
	Add Operation = iota
	Delete
)

type Update struct {
	Op   Operation
	Name string
	Addr string
}

type Builder interface {
	Build(Target) (Resolver, error)
	Scheme() string
}

type Resolver interface {
	Resolve(string) (Watcher, error)
}

type Watcher interface {
	Next() ([]*Update, error)
	Close()
}
