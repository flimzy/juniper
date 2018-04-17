package view

import (
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
			name: "no template",
			view: &view{templateDir: "/"},
			req:  httptest.NewRequest("GET", "/", nil),
			err:  "no template name provided",
		},
		{
			name: "file not found",
			view: &view{templateDir: ".", defTemplate: "oink"},
			req:  httptest.NewRequest("GET", "/", nil),
			err:  "open ./oink: no such file or directory",
		},
		{
			name:     "success",
			view:     &view{templateDir: "../test", defTemplate: "test.tmpl"},
			req:      httptest.NewRequest("GET", "/", nil),
			expected: `; defined templates are: "test.tmpl"`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := test.view.getTemplate(test.req)
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
			req:      stashRequest("GET", "/", nil, map[string]interface{}{"template": "bar"}),
			expected: "bar",
		},
		{
			name:     "invalid stash value",
			view:     &view{defTemplate: "foo"},
			req:      stashRequest("GET", "/", nil, map[string]interface{}{"template": 123}),
			expected: "foo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.view.templateName(test.req)
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
		header map[string][]string
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
			view:   &view{templateDir: "../test", defTemplate: "test.tmpl"},
			status: http.StatusOK,
			body:   "Test template",
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
