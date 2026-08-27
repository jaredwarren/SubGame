//go:build js

package save

import (
	"fmt"
	"io/fs"
	"syscall/js"
)

// localStorageBackend persists saves in the browser's localStorage (WASM builds).
type localStorageBackend struct{}

const localStorageKeyPrefix = "subgame:"

func localStorageObj() js.Value { return js.Global().Get("localStorage") }

func (localStorageBackend) read(name string) ([]byte, error) {
	v := localStorageObj().Call("getItem", localStorageKeyPrefix+name)
	if v.IsNull() {
		return nil, fmt.Errorf("save %q: %w", name, fs.ErrNotExist)
	}
	return []byte(v.String()), nil
}

func (localStorageBackend) write(name string, data []byte) error {
	localStorageObj().Call("setItem", localStorageKeyPrefix+name, string(data))
	return nil
}

func (localStorageBackend) remove(name string) error {
	localStorageObj().Call("removeItem", localStorageKeyPrefix+name)
	return nil
}

func (localStorageBackend) exists(name string) bool {
	return !localStorageObj().Call("getItem", localStorageKeyPrefix+name).IsNull()
}

var store storageBackend = localStorageBackend{}
