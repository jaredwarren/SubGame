//go:build !js

package save

import "os"

// fileStorage persists saves as files in the working directory (desktop builds).
type fileStorage struct{}

func (fileStorage) read(name string) ([]byte, error) { return os.ReadFile(name) }

// write is atomic: data lands in a temp file first, then replaces the target.
func (fileStorage) write(name string, data []byte) error {
	tmp := name + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, name)
}

func (fileStorage) remove(name string) error { return os.Remove(name) }

func (fileStorage) exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

var store storageBackend = fileStorage{}
