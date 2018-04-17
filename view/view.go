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
func New(dir, defTemplate string, funcMap map[string]interface{}) func(http.Handler) http.Handler {
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

func (v *view) templateName(r *http.Request) string {
	if tmpl, ok := GetStash(r)["template"].(string); ok {
		return tmpl
	}
	return v.defTemplate
}

func (v *view) render(w http.ResponseWriter, r *http.Request) {
	tmpl, err := v.getTemplate(r)
	if err != nil {
		httperr.HandleError(w, err)
		return
	}
	if e := tmpl.Execute(w, map[string]interface{}{}); e != nil {
		httperr.HandleError(w, err)
		return
	}
}

func (v *view) getTemplate(r *http.Request) (*template.Template, error) {
	if v.templateDir == "" {
		return nil, errors.New("template dir not defined")
	}
	name := v.templateName(r)
	if name == "" {
		return nil, errors.New("no template name provided")
	}
	return template.ParseFiles(v.templateDir + "/" + name)
}