// Package view provides middleware to add MVC-style View support to a Go web
// application.
package view

import (
	"errors"
	"html/template"
	"net/http"

	"github.com/flimzy/juniper/donewriter"
	"github.com/flimzy/juniper/httperr"
)

type view struct {
	templateDir string
	defTemplate string
	funcMap     map[string]interface{}
}

// New returns a new View middleware instance.
func New(dir, defTemplate string, funcMap template.FuncMap) func(http.Handler) http.Handler {
	v := &view{
		templateDir: dir,
		defTemplate: defTemplate,
		funcMap:     funcMap,
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
	if status, ok := stash[StashKeyStatus].(int); ok {
		w.WriteHeader(status)
	}
	if e := tmpl.Funcs(funcMap).ExecuteTemplate(w, tmplName, stash); e != nil {
		httperr.HandleError(w, err)
		return
	}
}

func (v *view) getTemplate(r *http.Request, name string) (*template.Template, error) {
	if v.templateDir == "" {
		return nil, errors.New("template dir not defined")
	}
	return template.New(name).Funcs(v.funcMap).ParseFiles(v.templateDir + "/" + name)
}
