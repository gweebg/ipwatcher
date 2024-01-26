package main

import (
	"github.com/gweebg/ipwatcher/internal/config"
	"log"
)

func main() {

	config.Init()

	v4 := "v4"
	v6 := "v6"
	f := "field"

	s := config.Source{
		Name: "mock",
		Url: config.SourceUrl{
			V4: &v4,
			V6: &v6,
		},
		Type:  "json",
		Field: &f,
	}
	err := config.AddSource(s)

	if err != nil {

		log.Print(err.Error())
	} else {
		log.Println("good")
	}
}
