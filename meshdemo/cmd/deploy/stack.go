package main

import (
	"fmt"
	"log"

	"github.com/gobuffalo/packr/v2"
)

func deploy() {
	box := packr.New("assets", "./assets")
	s, err := box.FindString("demo.yaml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
}