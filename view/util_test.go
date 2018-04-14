package view

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
)

func stashRequest(method, path string, body io.Reader, stash map[string]interface{}) *http.Request {
	req := httptest.NewRequest(method, path, body)
	ctx := context.WithValue(req.Context(), stashContextKey, Stash(stash))
	return req.WithContext(ctx)
}
