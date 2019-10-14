/*
 Package template is a utility wrapper around packr2 intended to simplify usage for CLI tools and simplify
 error handling boilerplate: errors are fatal and will exit the process. Don't use this wrapper if that's
 not the behavior that you want.
*/
package template

import (
	"text/template"

	"github.com/gobuffalo/packr/v2"

	"github.com/subfuzion/meshdemo/pkg/io"
)

type Template struct {
	Box *packr.Box
}

func New(box *packr.Box) *Template {
	return &Template{Box: box}
}

func (t *Template) ReadBytes(name string) []byte {
	b, err := t.Box.Find(name)
	if err != nil {
		io.Fatal(1, err)
	}
	return b
}

func (t *Template) Read(name string) string {
	s, err := t.Box.FindString(name)
	if err != nil {
		io.Fatal(1, err)
	}
	return s
}

func (t *Template) Parse(name string) *template.Template {
	tmpl := template.New(name)
	tmpl, err := tmpl.Parse(t.Read(name))
	if err != nil {
		io.Fatal(1, "Parsing %s: ", name, err)
	}
	return tmpl
}
