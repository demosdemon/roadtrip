package mapping

import (
	"fmt"
	"os"

	"github.com/demosdemon/roadtrip/pkg/utils"
)

type Provider map[string]string

// Clearenv takes a pointer receiver because it's faster
func (p *Provider) Clearenv() {
	*p = make(Provider, 0)
}

func (p Provider) Environ() []string {
	rv := make([]string, 0, len(p))
	for k, v := range p {
		rv = append(rv, fmt.Sprintf("%s=%s", k, v))
	}
	return rv
}

func (p Provider) ExpandEnv(key string) string {
	return os.Expand(key, p.Getenv)
}

func (p Provider) Getenv(key string) string {
	v, _ := p[key]
	return v
}

func (p Provider) LookupEnv(key string) (string, bool) {
	v, ok := p[key]
	return v, ok
}

func (p Provider) Setenv(key, value string) error {
	p[key] = value
	return nil
}

func (p Provider) Unsetenv(key string) error {
	delete(p, key)
	return nil
}

func New(environ []string) *Provider {
	rv := make(Provider, len(environ))
	for _, kv := range environ {
		k, _, v := utils.PartitionString(kv, "=")
		rv[k] = v
	}
	return &rv
}
