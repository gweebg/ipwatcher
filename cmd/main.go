package main

import (
	"github.com/gweebg/ipwatcher/internal/config"
	"github.com/gweebg/ipwatcher/internal/fetch"
	"github.com/gweebg/ipwatcher/internal/utils"
	"log"
)

func main() {

	config.Init()

	const version string = "v6"
	addr, err := fetch.RequestAddress(version)

	utils.Check(err, "")
	log.Println(addr)
}
