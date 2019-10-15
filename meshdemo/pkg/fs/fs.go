/*
Copyright © 2019 Tony Pujals <tpujals@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Package fs provides a few simple helper functions for file management
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
