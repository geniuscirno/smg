package registrator

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

// Endpoint endpoint of a service.
type Endpoint struct {
	Name    string      `json:"name"`
	Addr    string      `json:"addr"`
	Version string      `json:"version"`
	Meta    interface{} `json:"meta"`
}

// Target represents a terget for registrator.
// example:
//		etcd://localhost:2379
type Target struct {
	Scheme    string
	Authority string
	Endpoint  string
}

// Builder creates a client.
type Builder interface {
	Build(target Target) (Registrator, error)
	Scheme() string
}

//Registrator represents a registrator.
type Registrator interface {
	Register(string, *Endpoint) error
	Degister(string, *Endpoint) error
}
