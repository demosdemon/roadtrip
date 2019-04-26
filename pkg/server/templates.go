package server

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin/render"
	"github.com/spf13/afero"
	"html/template"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
)

const (
	templatesPath = "/templates/"
	layoutsPath   = "/templates/_layout/"
)

var (
	_ render.HTMLRender = (*templateEngine)(nil)
	_ render.Render     = (*templateRender)(nil)
)

type (
	templateEngine struct {
		*Server
		cacheMu sync.Mutex
		cache   map[string]*template.Template
	}

	templateRender struct {
		*templateEngine
		name string
		data interface{}

		cache []byte
	}

	TemplateError struct {
		Name string
		Data interface{}
		Err  error
	}
)

func (e *templateEngine) Instance(name string, data interface{}) render.Render {
	return &templateRender{
		templateEngine: e,
		name:           name,
		data:           data,
	}
}

func (e *templateEngine) loadTemplate(name string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"env":    e.Environment.Getenv,
		"static": e.staticURL,
	}

	var tpl *template.Template

	newTemplate := func(name, path string) error {
		data, err := afero.ReadFile(e.FileSystem, path)
		if err != nil {
			return &os.PathError{
				Op:   "loadTemplate",
				Path: path,
				Err:  err,
			}
		}

		if tpl == nil {
			tpl = template.New(name).Funcs(funcMap)
		} else {
			tpl = tpl.New(name)
		}

		tpl, err = tpl.Parse(string(data))
		if err != nil {
			return &os.PathError{
				Op:   "parseTemplate",
				Path: path,
				Err:  err,
			}
		}

		return nil
	}

	err := afero.Walk(e.FileSystem, layoutsPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		name := strings.TrimPrefix(path, layoutsPath)
		return newTemplate(name, path)
	})

	if err != nil {
		return nil, err
	}

	err = newTemplate(name, path.Join(templatesPath, name))
	return tpl, err
}

func (e *templateEngine) template(name string) (*template.Template, error) {
	if e.Debug {
		// skip cache
		return e.loadTemplate(name)
	}

	e.cacheMu.Lock()
	defer e.cacheMu.Unlock()

	if tpl, ok := e.cache[name]; ok {
		return tpl, nil
	}

	tpl, err := e.loadTemplate(name)
	if err != nil {
		return nil, err
	}

	if e.cache == nil {
		e.cache = make(map[string]*template.Template)
	}

	e.cache[name] = tpl
	return tpl, nil
}

func (e *templateEngine) clearCache() {
	e.cacheMu.Lock()
	e.cache = make(map[string]*template.Template)
	e.cacheMu.Unlock()
}

func (r *templateRender) Render(w http.ResponseWriter) error {
	if r.cache != nil {
		_, err := w.Write(r.cache)
		return err
	}

	tpl, err := r.template(r.name)
	if err != nil {
		return r.error(err)
	}

	var buf bytes.Buffer

	err = tpl.Execute(&buf, r.data)
	if err != nil {
		return r.error(err)
	}

	r.cache = buf.Bytes()
	_, err = w.Write(r.cache)

	return err
}

func (r *templateRender) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if _, ok := header["Content-Type"]; !ok {
		header.Set("Content-Type", "text/html; charset=utf-8")
	}
}

func (r *templateRender) error(err error) *TemplateError {
	return newTemplateError(r.name, r.data, err)
}

func (e *TemplateError) Error() string {
	return fmt.Sprintf("unable to render template %s with %v: %v", e.Name, e.Data, e.Err)
}

func newTemplateError(name string, data interface{}, err error) *TemplateError {
	return &TemplateError{
		Name: name,
		Data: data,
		Err:  err,
	}
}
