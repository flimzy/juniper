package httperr

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"
)

func TestStatusCoder(t *testing.T) {
	type scTest struct {
		name     string
		err      error
		expected int
	}
	tests := []scTest{
		{
			name:     "nil",
			expected: 0,
		},
		{
			name:     "Standard error",
			err:      errors.New("foo"),
			expected: 500,
		},
		// {
		// 	name:     "StatusCoder",
		// 	err:      kerrors.Status(400, "bad request"),
		// 	expected: 400,
		// },
	}
	for _, test := range tests {
		func(test scTest) {
			t.Run(test.name, func(t *testing.T) {
				result := StatusCode(test.err)
				if result != test.expected {
					t.Errorf("Unexpected result. Expected %d, got %d", test.expected, result)
				}
			})
		}(test)
	}
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name   string
		e      error
		err    string
		status int
		body   string
	}{
		{
			name:   "no error",
			e:      nil,
			status: http.StatusOK,
		},
		{
			name:   "standard error, text",
			e:      errors.New("foo"),
			status: http.StatusInternalServerError,
			body:   "Error 500: foo",
		},
		{
			name:   "status code error, text",
			e:      Wrap(http.StatusNotFound, errors.New("not found")),
			status: http.StatusNotFound,
			body:   "Error 404: not found",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := HandleError(w, test.e)
			res := w.Result()
			defer res.Body.Close()
			testy.Error(t, test.err, err)
			if test.status != res.StatusCode {
				t.Errorf("Unexpected status code: %d", res.StatusCode)
			}
			body, e := ioutil.ReadAll(res.Body)
			if e != nil {
				t.Fatal(e)
			}
			if d := diff.Text(test.body, string(body)); d != nil {
				t.Error(d)
			}
		})
	}
}
