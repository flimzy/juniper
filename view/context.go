package view

import (
	"context"
	"net/http"
)

// Stash represents per-request values which are passed to the template at the
// end of request processing.
type Stash map[string]interface{}

type contextKey struct {
	name string
}

// stashContextKey is a context key used to fetch the stash from a context. The
// returned value is of type Stash
var stashContextKey = &contextKey{"stash"}

// GetStash returns the stash stored in the request, or a nil map if no stash
// was found.
func GetStash(r *http.Request) Stash {
	if r == nil {
		return nil
	}
	stash, _ := r.Context().Value(stashContextKey).(Stash)
	return stash
}

// setStash creates an empty stash, and stores it in the passed request's
// context, then returns the request with the new context.
func setStash(r *http.Request) *http.Request {
	stash := Stash(make(map[string]interface{}))
	ctx := context.WithValue(r.Context(), stashContextKey, stash)
	return r.WithContext(ctx)
}
