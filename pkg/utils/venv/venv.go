package venv

import (
	"bufio"
	"github.com/spf13/afero"
	"log"
	"os"
	"path"

	"github.com/demosdemon/roadtrip/pkg/utils"
	"github.com/demosdemon/roadtrip/pkg/utils/venv/mapping"
	"github.com/demosdemon/roadtrip/pkg/utils/venv/passthru"
)

var (
	DefaultEnvironmentProvider EnvironmentProvider = passthru.Provider{}
)

type EnvironmentProvider interface {
	Clearenv()
	Environ() []string
	ExpandEnv(key string) string
	Getenv(key string) string
	LookupEnv(key string) (string, bool)
	Setenv(key, value string) error
	Unsetenv(key string) error
}

func CloneEnvironment(in EnvironmentProvider) (out EnvironmentProvider) {
	if in == nil {
		in = DefaultEnvironmentProvider
	}

	return mapping.New(in.Environ())
}

func ReadDotEnv(fs afero.Fs, p EnvironmentProvider) error {
	fp, err := locateDotEnv(fs)
	if err != nil {
		log.Printf("did not find a .env file: %v", err)
		return err
	}

	log.Printf("found .env file at %s", fp.Name())

	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		k, _, v := utils.PartitionString(scanner.Text(), "=")
		log.Printf("adding key %s to environment", k)
		_ = p.Setenv(k, v)
	}

	return scanner.Err()
}

func locateDotEnv(fs afero.Fs) (afero.File, error) {
	cwd := "/"
	for {
		dotenv := path.Join(cwd, ".env")
		fp, err := fs.Open(dotenv)
		if err == nil {
			return fp, nil
		}

		if !os.IsNotExist(err) {
			return nil, err
		}

		if cwd == "/" {
			return nil, os.ErrNotExist
		}

		cwd = path.Dir(cwd)
	}
}
