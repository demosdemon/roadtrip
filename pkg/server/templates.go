package server

import (
	"bytes"
	"github.com/gin-gonic/gin/render"
	"github.com/spf13/afero"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

const prefix = "/templates/_layout/"

var (
	_ render.HTMLRender = (*templateEngine)(nil)
	_ render.Render     = (*templateRender)(nil)
)

type (
	templateEngine struct {
		server *Server

		cacheMu sync.Mutex
		cache   map[string]*template.Template
	}

	templateRender struct {
		engine *templateEngine
		name   string
		data   interface{}
	}
)

func (e *templateEngine) Instance(name string, data interface{}) render.Render {
	return &templateRender{
		engine: e,
		name:   name,
		data:   data,
	}
}

func (e *templateEngine) template(name string) (*template.Template, error) {
	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()

	if tpl, ok := e.cache[name]; ok {
		return tpl, nil
	}

	funcMap := template.FuncMap{
		"env": e.server.Environment.Getenv,
	}

	var tpl *template.Template

	err := afero.Walk(e.server.FileSystem, prefix, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		name := strings.TrimPrefix(path, prefix)
		if tpl == nil {
			tpl = template.New(name).Funcs(funcMap)
		} else {
			tpl = tpl.New(name)
		}

		data, err := afero.ReadFile(e.server.FileSystem, path)
		if err != nil {
			return err
		}

		tpl, err = tpl.Parse(string(data))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	data, err := afero.ReadFile(e.server.FileSystem, "/templates/"+name)
	if err != nil {
		return nil, err
	}

	if tpl == nil {
		tpl = template.New(name).Funcs(funcMap)
	} else {
		tpl = tpl.New(name)
	}

	tpl, err = tpl.Parse(string(data))
	if err != nil {
		return nil, err
	}

	if e.cache == nil {
		e.cache = make(map[string]*template.Template)
	}

	e.cache[name] = tpl
	return tpl, nil
}

func (r *templateRender) Render(w http.ResponseWriter) error {
	tpl, err := r.engine.template(r.name)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, r.data)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, buf)
	return err
}

func (*templateRender) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if _, ok := header["Content-Type"]; !ok {
		header.Set("Content-Type", "text/html; charset=utf-8")
	}
}
