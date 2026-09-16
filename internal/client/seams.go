package client

import "sync"

var seamMu sync.RWMutex //nolint:gochecknoglobals // guards package-level test seams

//nolint:ireturn // copies a typed seam value; T is not an interface policy
func loadSeam[T any](slot *T) T {
	seamMu.RLock()

	value := *slot

	seamMu.RUnlock()

	return value
}

func swapSeam[T any](slot *T, next T) func() {
	seamMu.Lock()

	orig := *slot
	*slot = next

	seamMu.Unlock()

	return func() {
		seamMu.Lock()
		*slot = orig
		seamMu.Unlock()
	}
}
