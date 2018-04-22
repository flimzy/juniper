package view

const (
	// StashKeyRequest is a stash key used to store the HTTP request, for use
	// by templates.
	StashKeyRequest = "_req"
	// StashKeyFuncMap is a stash key used to store a template.FuncMap, which is
	// used to override or augment the default funcmap.
	StashKeyFuncMap = "_funcs"
	// StashKeyTemplate is a stash key used to store the name of the template
	// to use for rendering the request.
	StashKeyTemplate = "_template"
	// StashKeyStatus is a stash key which, if set to an int, defines the status
	// code to be returned when rendering the template.
	StashKeyStatus = "_status"
)

const (
	// DefaultContentType is the default Content-Type for all responses which do
	// not already have their Content-Type header set by view rendering time.
	DefaultContentType = "text/html; charset=utf-8"
)
