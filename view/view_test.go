package view

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"
)

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		name     string
		view     *view
		req      *http.Request
		tmplName string
		expected string
		err      string
	}{
		{
			name: "no template dir",
			view: &view{},
			req:  httptest.NewRequest("GET", "/", nil),
			err:  "template dir not defined",
		},
		{
			name:     "file not found",
			view:     &view{templateDir: "."},
			req:      httptest.NewRequest("GET", "/", nil),
			tmplName: "oink",
			err:      `failed to parse template "oink": open ./oink: no such file or directory`,
		},
		{
			name:     "success",
			view:     &view{templateDir: "test"},
			req:      httptest.NewRequest("GET", "/", nil),
			tmplName: "test.tmpl",
			expected: `; defined templates are: "test.tmpl"`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := test.view.getTemplate(test.req, test.tmplName)
			testy.Error(t, test.err, err)
			if tmpl.DefinedTemplates() != test.expected {
				t.Errorf("Unexpected result: %s", tmpl.DefinedTemplates())
			}
		})
	}
}

func TestTemplateName(t *testing.T) {
	tests := []struct {
		name     string
		view     *view
		req      *http.Request
		err      string
		expected string
	}{
		{
			name:     "Default",
			view:     &view{defTemplate: "foo"},
			expected: "foo",
		},
		{
			name:     "from stash",
			view:     &view{defTemplate: "foo"},
			req:      stashRequest("GET", "/", nil, map[string]interface{}{StashKeyTemplate: "bar"}),
			expected: "bar",
		},
		{
			name:     "invalid stash value",
			view:     &view{defTemplate: "foo"},
			req:      stashRequest("GET", "/", nil, map[string]interface{}{StashKeyTemplate: 123}),
			expected: "foo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.view.templateName(test.req)
			testy.Error(t, test.err, err)
			if result != test.expected {
				t.Errorf("Unexpected result: %s", result)
			}
		})
	}
}

func TestRender(t *testing.T) {
	tests := []struct {
		name   string
		view   *view
		req    *http.Request
		status int
		header http.Header
		body   string
	}{
		{
			name:   "no template defined",
			view:   &view{templateDir: "."},
			status: http.StatusInternalServerError,
			body:   "Error 500: no template name provided",
		},
		{
			name:   "success",
			view:   &view{templateDir: "test", defTemplate: "test.tmpl"},
			req:    setStash(httptest.NewRequest("GET", "/", nil)),
			status: http.StatusOK,
			header: http.Header{
				"Content-Type": []string{DefaultContentType},
			},
			body: "Test template",
		},
		{
			name: "with stash",
			view: &view{templateDir: "test", defTemplate: "hello.tmpl"},
			req: func() *http.Request {
				r := httptest.NewRequest("GET", "/", nil)
				r = setStash(r)
				stash := GetStash(r)
				stash["Name"] = "Gregory"
				return r
			}(),
			status: http.StatusOK,
			body:   "Hello, Gregory!",
		},
		{
			name:   "request details",
			view:   &view{templateDir: "test", defTemplate: "req.tmpl"},
			req:    setStash(httptest.NewRequest("GET", "/foo/bar.html", nil)),
			status: http.StatusOK,
			body:   "GET /foo/bar.html from 192.0.2.1:1234",
		},
		{
			name: "with function",
			view: &view{templateDir: "test", defTemplate: "foo.tmpl",
				funcMap: map[string]interface{}{"foo": func() string { return "foo!" }},
			},
			req:    setStash(httptest.NewRequest("GET", "/foo/bar.html", nil)),
			status: http.StatusOK,
			body:   "Foo? foo!",
		},
		{
			name: "with custom function",
			view: &view{templateDir: "test", defTemplate: "foo.tmpl",
				funcMap: map[string]interface{}{"foo": func() string { return "foo!" }},
			},
			req: func() *http.Request {
				r := setStash(httptest.NewRequest("GET", "/foo/bar.html", nil))
				stash := GetStash(r)
				stash[StashKeyFuncMap] = map[string]interface{}{
					"foo": func() string { return "no foo :(" },
				}
				return r
			}(),
			status: http.StatusOK,
			body:   "Foo? no foo :(",
		},
		{
			name:   "with includes",
			view:   &view{templateDir: "test", defTemplate: "lib.tmpl", entryPoint: "base.tmpl", includes: []string{"test/lib"}},
			req:    setStash(httptest.NewRequest("GET", "/", nil)),
			status: http.StatusOK,
			body:   "before\n\nincluded\n\nafter",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			test.view.render(rec, test.req)
			res := rec.Result()
			defer res.Body.Close()
			if res.StatusCode != test.status {
				t.Errorf("Unexpected status: %d", res.StatusCode)
			}
			if test.header != nil {
				if d := diff.Interface(test.header, res.Header); d != nil {
					t.Error(d)
				}
			}
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if d := diff.Text(test.body, string(body)); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name    string
		conf    Config
		handler http.Handler
		req     *http.Request
		status  int
		header  http.Header
		body    string
	}{
		{
			name: "custom status code",
			conf: Config{TemplateDir: "test", DefaultTemplate: "test.tmpl"},
			handler: http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				stash := GetStash(r)
				stash[StashKeyStatus] = 600
			}),
			status: 600,
			body:   "Test template",
		},
		{
			name: "default status code",
			conf: Config{TemplateDir: "test", DefaultTemplate: "test.tmpl"},
			handler: http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
				// Do nothing
			}),
			status: http.StatusOK,
			header: http.Header{
				"Content-Type": []string{DefaultContentType},
			},
			body: "Test template",
		},
		{
			name: "already written",
			conf: Config{TemplateDir: "test", DefaultTemplate: "test.tmpl"},
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(300)
			}),
			status: 300,
		},
		{
			name: "custom headers",
			conf: Config{TemplateDir: "test", DefaultTemplate: "test.tmpl"},
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				w.Header().Set("X-Foo", "bar")
				// Do nothing
			}),
			status: http.StatusOK,
			header: http.Header{
				"Content-Type": []string{"text/plain"},
				"X-Foo":        []string{"bar"},
			},
			body: "Test template",
		},
		{
			name: "funcMaps",
			conf: Config{TemplateDir: "test", DefaultTemplate: "foo.tmpl",
				FuncMaps: []template.FuncMap{{"foo": func() string { return "FoO" }}},
			},
			handler: http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
				// Do nothing
			}),
			status: http.StatusOK,
			body:   "Foo? FoO",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := New(test.conf)(test.handler)
			w := httptest.NewRecorder()
			req := test.req
			if req == nil {
				req = httptest.NewRequest(http.MethodGet, "/", nil)
			}
			handler.ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != test.status {
				t.Errorf("Unexpected status code: %d", res.StatusCode)
			}
			if test.header != nil {
				if d := diff.Interface(test.header, res.Header); d != nil {
					t.Error(d)
				}
			}
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if d := diff.Text(test.body, string(body)); d != nil {
				t.Error(d)
			}
		})
	}
}
