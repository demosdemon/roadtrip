package passthru

import "os"

type Provider struct{}

func (Provider) Clearenv() { os.Clearenv() }

func (Provider) Environ() []string { return os.Environ() }

func (Provider) ExpandEnv(key string) string { return os.ExpandEnv(key) }

func (Provider) Getenv(key string) string { return os.Getenv(key) }

func (Provider) LookupEnv(key string) (string, bool) { return os.LookupEnv(key) }

func (Provider) Setenv(key, value string) error { return os.Setenv(key, value) }

func (Provider) Unsetenv(key string) error { return os.Unsetenv(key) }
