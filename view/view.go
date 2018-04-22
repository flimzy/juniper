// Package view provides middleware to add MVC-style View support to a Go web
// application.
package view

import (
	"html/template"
	"net/http"

	"github.com/pkg/errors"

	"github.com/flimzy/juniper/donewriter"
	"github.com/flimzy/juniper/httperr"
)

type view struct {
	templateDir string
	entryPoint  string
	defTemplate string
	funcMap     map[string]interface{}
	includes    []string
}

type Config struct {
	// TemplateDir is the root dir where templates can be found.
	TemplateDir string
	// DefaultTemplate is the name of the default template (to be found in
	// TemplateDir) which will be used if stash[StashKeyTemplate] is undefined.
	DefaultTemplate string
	// FuncMap defines the default FuncMap to be passed when templates are
	// parsed. This may be overwritten or augmented with stash[StashKeyFuncMap]
	FuncMap template.FuncMap
	// Includes is zero or more paths to include when parsing all templates.
	// This can be used to define global templates or components
	Includes []string
	// EntryPoint defines the template that is executed by the
	// template.ExecuteTemplate call. This will typically be a basic HTML
	// template, which is populated by calls to the specific template. If unset,
	// falls back to the template name. This value may be overwridden per request
	// by the stash[StashKeyEntryPoint] value
	EntryPoint string
}

// New returns a new View middleware instance. It accepts the following arguments:
//
// dir:         The root dir where templates are to be found
// defTemplate:
func New(c Config) func(http.Handler) http.Handler {
	v := &view{
		templateDir: c.TemplateDir,
		entryPoint:  c.EntryPoint,
		defTemplate: c.DefaultTemplate,
		funcMap:     c.FuncMap,
		includes:    c.Includes,
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			w := donewriter.New(rw)
			r = setStash(r)
			next.ServeHTTP(w, r)
			if w.Done() {
				return
			}
			v.render(w, r)
		})
	}
}

func (v *view) templateName(r *http.Request) (string, error) {
	if tmpl, ok := GetStash(r)[StashKeyTemplate].(string); ok {
		return tmpl, nil
	}
	if v.defTemplate != "" {
		return v.defTemplate, nil
	}
	return "", errors.New("no template name provided")
}

func (v *view) render(w http.ResponseWriter, r *http.Request) {
	tmplName, err := v.templateName(r)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}
	tmpl, err := v.getTemplate(r, tmplName)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}
	stash := GetStash(r)
	stash[StashKeyRequest] = r
	funcMap := v.funcMap
	if m, ok := stash[StashKeyFuncMap]; ok {
		var fm template.FuncMap
		switch t := m.(type) {
		case template.FuncMap:
			fm = t
		case map[string]interface{}:
			fm = t
		}
		for key, val := range fm {
			funcMap[key] = val
		}
	}
	if _, ok := w.Header()["Content-Type"]; !ok {
		w.Header().Set("Content-Type", DefaultContentType)
	}
	if status, ok := stash[StashKeyStatus].(int); ok {
		w.WriteHeader(status)
	}
	entryPoint := v.entryPoint
	if ep, ok := stash[StashKeyEntryPoint].(string); ok {
		entryPoint = ep
	}
	if entryPoint == "" {
		entryPoint = tmplName
	}

	if e := tmpl.Funcs(funcMap).ExecuteTemplate(w, entryPoint, stash); e != nil {
		httperr.HandleError(w, err)
		return
	}
}

func (v *view) getTemplate(r *http.Request, name string) (*template.Template, error) {
	if v.templateDir == "" {
		return nil, errors.New("template dir not defined")
	}
	t := template.New("")
	t.Funcs(v.funcMap)
	if _, err := t.ParseFiles(v.templateDir + "/" + name); err != nil {
		return nil, errors.Wrapf(err, "failed to parse template %q", name)
	}
	for _, libPath := range v.includes {
		if _, err := t.ParseGlob(libPath + "/*"); err != nil {
			return nil, errors.Wrapf(err, "failed to parse include path '%s'", libPath)
		}
	}
	return t, nil
}
