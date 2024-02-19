package main

import (
	"flag"
	"github.com/gweebg/ipwatcher/internal/config"
	"github.com/gweebg/ipwatcher/internal/database"
	"github.com/gweebg/ipwatcher/internal/utils"
	"github.com/gweebg/ipwatcher/internal/watcher"
	"log"
)

func main() {

	configFlags := map[string]interface{}{}

	configFlags["version"] = flag.String(
		"version",
		"v4",
		"version of the IP protocol, supports 'v4' | 'v6'",
	)

	configFlags["config"] = flag.String(
		"config",
		"config.yml",
		"path to the configuration file",
	)

	configFlags["exec"] = flag.Bool(
		"exec",
		true,
		"enable execution of configuration defined actions",
	)

	configFlags["api"] = flag.Bool(
		"api",
		false,
		"expose an api with information relative to the changes",
	)

	configFlags["notify"] = flag.Bool(
		"notify",
		false,
		"enable email notifications, must be configured in settings",
	)

	configFlags["quiet"] = flag.Bool(
		"quiet",
		false,
		"run application quietly, sets log level to Info instead of Debug",
	)

	flag.Parse()

	version := configFlags["version"].(*string)
	if version == nil || (*version != "v4" && *version != "v6") {
		log.Fatalf("flag 'version' must be either 'v4' or 'v6', not '%v'\n", *version)
	}

	config.Init(configFlags)
	watcher.InitLogger()

	database.ConnectDatabase()
	db := database.GetDatabase()

	err := db.AutoMigrate(&database.AddressEntry{})
	utils.Check(err, "could not run database AutoMigrate")

	w := watcher.NewWatcher()
	w.Watch()
}
