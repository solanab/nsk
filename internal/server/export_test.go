package server

import (
	"net/http"
	"time"
)

const (
	PathHealth        = pathHealth
	PathWhoAmI        = pathWhoAmI
	PathMe            = pathMe
	PathPosts         = pathPosts
	PathSearch        = pathSearch
	PathCategories    = pathCategories
	PathNotifications = pathNotifications
)

// HeaderTimeout is ListenAndServe ReadHeaderTimeout.
func HeaderTimeout(srv *Server) time.Duration {
	if srv.http == nil {
		return 0
	}

	return srv.http.ReadHeaderTimeout
}

// WriteJSONForTest exposes writeJSON for encode-error coverage.
func WriteJSONForTest(writer http.ResponseWriter, status int, value any) {
	writeJSON(writer, status, value)
}

// FailWriter is a ResponseWriter whose Write fails.
type FailWriter struct {
	http.ResponseWriter

	Err error
}

// Write returns Err.
func (writer FailWriter) Write([]byte) (int, error) {
	return 0, writer.Err
}
