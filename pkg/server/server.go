package server

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"

	"github.com/demosdemon/roadtrip/pkg/utils/venv"
)

type (
	Server struct {
		Output      io.Writer
		Environment venv.EnvironmentProvider
		FileSystem  afero.Fs

		initOnce sync.Once
	}

	flusher interface {
		Flush() error
	}
)

func (s *Server) init() {
	s.initOnce.Do(func() {
		if s.Output == nil {
			s.Output = os.Stderr
		}

		if s.Environment == nil {
			env := venv.DefaultEnvironmentProvider
			env = venv.CloneEnvironment(env)

			s.Environment = env
		}

		if s.FileSystem == nil {
			fs := afero.NewOsFs()
			fs = afero.NewReadOnlyFs(fs)

			if cwd, err := os.Getwd(); err == nil {
				fs = afero.NewBasePathFs(fs, cwd)
			}

			s.FileSystem = fs
		}

		_ = venv.ReadDotEnv(s.FileSystem, s.Environment)
	})
}

func (s *Server) Handler() (http.Handler, error) {
	s.init()
	engine := gin.New()
	engine.Use(
		gin.ErrorLogger(),
		gin.LoggerWithWriter(s.Output),
		gin.RecoveryWithWriter(s.Output),
	)
	engine.RedirectTrailingSlash = true
	engine.RedirectFixedPath = true
	engine.HandleMethodNotAllowed = true
	engine.ForwardedByClientIP = true

	engine.HTMLRender = &templateEngine{
		server: s,
	}

	engine.GET("/", s.index)
	engine.StaticFS("/static", s.static())
	return engine, nil
}

func (s *Server) Listener() (net.Listener, error) {
	s.init()

	if socket, ok := s.Environment.LookupEnv("SOCKET"); ok {
		return net.Listen("unix", socket)
	}

	if port, ok := s.Environment.LookupEnv("PORT"); ok {
		return net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	}

	return net.Listen("tcp", "localhost:9000")
}

func (s *Server) index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":   "index",
		"content": "Hello, World!",
	})
}

func (s *Server) Flush() error {
	if f, ok := s.Output.(flusher); ok {
		return f.Flush()
	}
	return nil
}

func (s *Server) static() http.FileSystem {
	s.init()
	fs := afero.NewBasePathFs(s.FileSystem, "/static")
	return afero.NewHttpFs(fs)
}
