package httperr

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// copies from github.com/pkg/errors package
type causer interface {
	Cause() error
}

type statusError struct {
	error
	status int
}

var _ error = &statusError{}
var _ causer = &statusError{}
var _ statusCoder = &statusError{}

func (e *statusError) Cause() error {
	return e.error
}

func (e *statusError) StatusCode() int {
	return e.status
}

// Wrap bundles an existing error with a status code.
func Wrap(status int, err error) error {
	return &statusError{
		error:  err,
		status: status,
	}
}

// Wrapf bundles an existing error with a status code and adds a message.
func Wrapf(status int, err error, fmt string, args ...interface{}) error {
	return Wrap(status, errors.Wrapf(err, fmt, args...))
}

// New returns a new status-embedded error.
func New(status int, msg string) error {
	return Wrap(status, errors.New(msg))
}

// Errorf returns a new status-embedded error with formatting.
func Errorf(status int, fmt string, args ...interface{}) error {
	return Wrap(status, errors.Errorf(fmt, args...))
}

type statusCoder interface {
	StatusCode() int
}

// StatusCode returns the HTTP status code embedded in the error, or 500
// (internal server error), if there was no specified status code.  If err is
// nil, StatusCode returns 0.
//
// This method uses the statusCoder interface, which is not exported by this
// package, but is considered part of the stable public API.  Driver
// implementations are expected to return errors which conform to this
// interface.
//
//  type statusCoder interface {
//      StatusCode() int
//  }
func StatusCode(err error) int {
	if err == nil {
		return 0
	}
	if coder, ok := err.(statusCoder); ok {
		return coder.StatusCode()
	}
	return http.StatusInternalServerError
}

// HandleError serves an error response if e is non-nil. If the error embeds a
// status code via the statusCoder interface, If the response cannot
// be written, for instance if the response has already been sent, an error is
// returned. If e is nil, this function is a no-op.
func HandleError(w http.ResponseWriter, e error) error {
	if e == nil {
		return nil
	}
	status := StatusCode(e)
	w.WriteHeader(status)
	_, err := fmt.Fprintf(w, "Error %d: %s", status, e)
	return err
}
