package view

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flimzy/diff"
)

func TestGetStash(t *testing.T) {
	tests := []struct {
		name     string
		req      *http.Request
		expected Stash
	}{
		{
			name: "nil request",
		},
		{
			name: "no stash",
			req:  httptest.NewRequest("GET", "/", nil),
		},
		{
			name: "stash found",
			req: func() *http.Request {
				stash := Stash(map[string]interface{}{"foo": "bar"})
				req := httptest.NewRequest("GET", "/", nil)
				ctx := context.WithValue(req.Context(), StashContextKey, stash)
				return req.WithContext(ctx)
			}(),
			expected: map[string]interface{}{"foo": "bar"},
		},
		{
			name: "wrong type",
			req: func() *http.Request {
				stash := map[string]interface{}{"foo": "bar"}
				req := httptest.NewRequest("GET", "/", nil)
				ctx := context.WithValue(req.Context(), StashContextKey, stash)
				return req.WithContext(ctx)
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := GetStash(test.req)
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
