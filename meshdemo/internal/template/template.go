/*
 Package template is a utility wrapper around packr2 intended to simplify usage for CLI tools and simplify
 error handling boilerplate: errors are fatal and will exit the process. Don't use this wrapper if that's
 not the behavior that you want.

 This packages is hardcoded to uses a top-level "_templates" directory as the base path for templates.
 */
package template

import (
	"github.com/gobuffalo/packr/v2"
	"github.com/subfuzion/meshdemo/pkg/io"
)

var box *packr.Box

func init() {
	box = packr.New("templates", "_templates")
}

func ReadBytes(name string) []byte {
	b, err := box.Find(name)
	if err != nil {
		io.Fatal(1, err)
	}
	return b
}

func Read(name string) string {
	s, err := box.FindString(name)
	if err != nil {
		io.Fatal(1, err)
	}
	return s
}
