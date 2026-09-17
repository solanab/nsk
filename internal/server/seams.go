package server

import (
	"net/http"
	"sync"
)

var listenMu sync.RWMutex //nolint:gochecknoglobals // guards listen/close seams

// StubListenHTTP replaces net/http Server.ListenAndServe.
func StubListenHTTP(fn func(*http.Server) error) func() {
	listenMu.Lock()
	orig := listenHTTP
	listenHTTP = fn
	listenMu.Unlock()

	return func() {
		listenMu.Lock()
		listenHTTP = orig
		listenMu.Unlock()
	}
}

// StubCloseHTTP replaces net/http Server.Close.
func StubCloseHTTP(fn func(*http.Server) error) func() {
	listenMu.Lock()
	orig := closeHTTP
	closeHTTP = fn
	listenMu.Unlock()

	return func() {
		listenMu.Lock()
		closeHTTP = orig
		listenMu.Unlock()
	}
}

func currentListen() func(*http.Server) error {
	listenMu.RLock()

	listen := listenHTTP

	listenMu.RUnlock()

	return listen
}

func currentClose() func(*http.Server) error {
	listenMu.RLock()

	closeFn := closeHTTP

	listenMu.RUnlock()

	return closeFn
}
