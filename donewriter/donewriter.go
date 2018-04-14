// Package donewriter provides a simple wrapper around an http.ResponseWriter to
// track when a response has been sent.  To use it, call the WrapWriter
// middleware early in your middleware stack. Then in other middlewares or
// handlers, you can use the WriterIsDone method to check the status.
//
//  func main() {
//      r := chi.NewRouter()
//
//      r.Use(donewriter.WrapWriter)
//      // and other middlewares
//
//      r.Get("/", func(w http.ResponseWriter, r *http.Request) {
//          if done, _ := donewriter.WriterIsDone(w); done {
//              // Nothing to do, a response was already sent
//              return
//          }
//
//          // Normal operation here...
//
//      })
//  }
package donewriter

import (
	"errors"
	"net/http"
)

type DoneWriter interface {
	http.ResponseWriter
	Done() bool
}

// doneWriter is an http.ResponseWriter which tracks its write state.
type doneWriter struct {
	http.ResponseWriter
	done bool
}

var _ http.ResponseWriter = &doneWriter{}

// New returns a new DoneWriter instance which wraps rw.
func New(rw http.ResponseWriter) DoneWriter {
	return &doneWriter{ResponseWriter: rw}
}

func (w *doneWriter) Done() bool {
	return w.done
}

// WriteHeader wraps the underlying WriteHeader method.
func (w *doneWriter) WriteHeader(status int) {
	w.done = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *doneWriter) Write(b []byte) (int, error) {
	w.done = true
	return w.ResponseWriter.Write(b)
}

// WriterIsDone returns true if a response has been written. An error is
// returned if the underlying writer is not a DoneWriter.
func WriterIsDone(w http.ResponseWriter) (bool, error) {
	if dw, ok := w.(DoneWriter); ok {
		return dw.Done(), nil
	}
	return false, errors.New("not a DoneWriter")
}

// WrapWriter is an http middleware which wraps the standard http.ResponseWriter
// with a DoneWriter.  Subsequent middlewares or handlers should use the
// WriterIsDone method to check the status.
func WrapWriter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&doneWriter{ResponseWriter: w}, r)
	})
}
