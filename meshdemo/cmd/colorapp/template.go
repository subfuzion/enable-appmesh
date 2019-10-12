package main

import (
	"github.com/gobuffalo/packr/v2"

	"github.com/subfuzion/meshdemo/internal/template"
)

var tpl *template.Template

func init() {
	// use packr2 to compile _templates into the executable
	tpl = template.New(packr.New("templates", "_templates"))
}
