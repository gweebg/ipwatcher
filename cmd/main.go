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

	//if len(os.Args) > 1 {
	//
	//	command := os.Args[1]
	//	if command == "add-source" {
	//		// ! Execute CLI for source addition.
	//		// ! Maybe use a flag set to separate flags.
	//		// ! source := flag.NewFlagSet("add-source", flag.ExitOnError)
	//	}
	//
	//	return
	//}

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
		false,
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

	flag.Parse()

	version := configFlags["version"].(*string)
	if version == nil || (*version != "v4" && *version != "v6") {
		log.Fatalf("flag 'version' must be either 'v4' or 'v6', not '%v'\n", *version)
	}

	config.Init(configFlags)

	database.ConnectDatabase()
	db := database.GetDatabase()

	err := db.AutoMigrate(&database.AddressEntry{})
	utils.Check(err, "could not run database AutoMigrate")

	w := watcher.NewWatcher()
	w.Watch()
}
