package view

import "net/http"

// Stash represents per-request values which are passed to the template at the
// end of request processing.
type Stash map[string]interface{}

type contextKey struct {
	name string
}

// StashContextKey is a context key used to fetch the stash from a context. The
// returned value is of type Stash
var StashContextKey = &contextKey{"stash"}

// GetStash returns the stash stored in the request, or a nil map if no stash
// was found.
func GetStash(r *http.Request) Stash {
	if r == nil {
		return nil
	}
	stash, _ := r.Context().Value(StashContextKey).(Stash)
	return stash
}
