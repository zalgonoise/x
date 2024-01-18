package mapping

import "sync"

// SyncField is a decorator for a Field type (either a *Table or *Index instance), enabling it for use in
// concurrent operations.
//
// It uses a *sync.RWMutex to lock read and write operations, accordingly.
type SyncField[K comparable, T any] struct {
	field Field[K, T]
	mu    *sync.RWMutex
}

// Get fetches the value in a mapping Field for a given key. If the value does not exist, the Field's
// configured zero value is returned. A boolean value is also returned to highlight whether accessing the key was
// successful or not.
func (f SyncField[K, T]) Get(key K) (T, bool) {
	f.mu.RLock()
	value, ok := f.field.Get(key)
	f.mu.RUnlock()

	return value, ok
}

// Set replaces the value of a certain key in the map, or it adds it if it does not exist. The returned boolean value
// represents whether the key is new in the mapping Field or not.
func (f SyncField[K, T]) Set(key K, setter Setter[T]) bool {
	f.mu.Lock()
	ok := f.field.Set(key, setter)
	f.mu.Unlock()

	return ok
}
