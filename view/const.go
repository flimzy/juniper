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
)
