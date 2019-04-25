package venv

import (
	"bufio"
	"errors"
	"os"
	"path"
	"syscall"

	"github.com/spf13/afero"

	"github.com/demosdemon/roadtrip/pkg/utils"
	"github.com/demosdemon/roadtrip/pkg/utils/venv/mapping"
	"github.com/demosdemon/roadtrip/pkg/utils/venv/passthru"
)

var (
	ErrCrossedDeviceBoundary = errors.New("crossed device boundary")

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
		return err
	}

	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		k, _, v := utils.PartitionString(scanner.Text(), "=")
		_ = p.Setenv(k, v)
	}

	return scanner.Err()
}

func locateDotEnv(fs afero.Fs) (afero.File, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cwdInfo, err := fs.Stat(cwd)
	if err != nil {
		return nil, err
	}

	statObj, checkDevice := cwdInfo.Sys().(*syscall.Stat_t)

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
		if checkDevice {
			info, err := fs.Stat(cwd)
			if err != nil {
				return nil, err
			}

			if parentStatObj, ok := info.Sys().(*syscall.Stat_t); ok {
				if parentStatObj.Dev == statObj.Dev {
					continue
				}

				err := os.PathError{
					Op:   "locateDotEnv",
					Path: cwd,
					Err:  ErrCrossedDeviceBoundary,
				}
				return nil, &err
			}
		}
	}
}
