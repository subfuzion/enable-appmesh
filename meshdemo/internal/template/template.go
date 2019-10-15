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
// Package template is a utility wrapper around packr2 intended to simplify usage for CLI tools and simplify
// error handling boilerplate: errors are fatal and will exit the process. Don't use this wrapper if that's
// not the behavior that you want.
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
