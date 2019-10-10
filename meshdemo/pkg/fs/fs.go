package fs

import "os"

// EnsureFile will create the named file if it doesn't exist
func EnsureFile(name string) error {
	f, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

// CreateFileExclusive will create the named file if it doesn't exist,
// but returns an error if it does
func CreateFileExclusive(name string) error {
	f, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}
