package server

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"

	"github.com/demosdemon/roadtrip/pkg/utils/venv"
)

type (
	Server struct {
		context.Context
		Debug       bool
		Output      io.Writer
		Environment venv.EnvironmentProvider
		FileSystem  afero.Fs

		initOnce sync.Once

		staticURLMu    sync.Mutex
		staticURLCache map[string]string
	}

	flusher interface {
		Flush() error
	}
)

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

	tplEngine := &templateEngine{
		Server: s,
	}
	registerHUP(s, tplEngine)
	engine.HTMLRender = tplEngine

	engine.GET("/", s.index)
	engine.StaticFS("/static", s.static())
	engine.POST("/location", func(c *gin.Context) {
		var input struct {
			Latitude  float64 `form:"lat" json:"lat" xml:"Latitude" binding:"required"`
			Longitude float64 `form:"lng" json:"lng" xml:"Longitude" binding:"required"`
		}

		err := c.Bind(&input)
		if err != nil {
			_ = c.Error(err)
			return
		}

		c.JSON(http.StatusAccepted, input)
	})
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

func (s *Server) Flush() error {
	if f, ok := s.Output.(flusher); ok {
		return f.Flush()
	}
	return nil
}

func (s *Server) init() {
	s.initOnce.Do(func() {
		registerHUP(s, s)

		if s.Context == nil {
			log.Print("received a nil context, defaulting to context.Background()")
			s.Context = context.Background()
		}

		if s.Output == nil {
			log.Print("received a nil Writer, defaulting to os.Stdout")
			s.Output = os.Stdout
		}

		if s.Environment == nil {
			log.Print("received a nil Environment, cloning the DefaultEnvironmentProvider")
			env := venv.DefaultEnvironmentProvider
			env = venv.CloneEnvironment(env)

			s.Environment = env
		}

		if s.FileSystem == nil {
			log.Print("received a nil Filesystem, using a read-only filesystem")
			fs := afero.NewOsFs()
			fs = afero.NewReadOnlyFs(fs)

			if cwd, err := os.Getwd(); err == nil {
				log.Printf("changing filesystem root to %q", cwd)
				fs = afero.NewBasePathFs(fs, cwd)
			} else {
				log.Printf("unable to determine PWD: %v", err)
			}

			s.FileSystem = fs
		}

		s.staticURLCache = make(map[string]string)

		_ = venv.ReadDotEnv(s.FileSystem, s.Environment)
	})
}

func (s *Server) index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":   "index",
		"content": "Hello, World!",
	})
}

func (s *Server) static() http.FileSystem {
	fs := afero.NewBasePathFs(s.FileSystem, "/static")
	return afero.NewHttpFs(fs)
}

func (s *Server) staticURL(name string) (string, error) {
	staticPath := path.Join("/static", name)

	s.staticURLMu.Lock()
	defer s.staticURLMu.Unlock()

	if v, ok := s.staticURLCache[name]; ok {
		return v, nil
	}

	fp, err := s.FileSystem.Open(staticPath)
	if err != nil {
		return "", err
	}
	defer fp.Close()

	hash := md5.New()
	_, err = io.Copy(hash, fp)
	if err != nil {
		return "", err
	}

	md5sum := hex.EncodeToString(hash.Sum(nil))
	rv := fmt.Sprintf("%s?t=%s", staticPath, md5sum)

	if !s.Debug {
		s.staticURLCache[name] = rv
	}

	return rv, nil
}

func (s *Server) clearCache() {
	s.staticURLMu.Lock()
	s.staticURLCache = make(map[string]string)
	s.staticURLMu.Unlock()
}
