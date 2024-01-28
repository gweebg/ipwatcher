package main

import (
	"flag"
	"github.com/gweebg/ipwatcher/internal/config"
	"github.com/gweebg/ipwatcher/internal/database"
	"github.com/gweebg/ipwatcher/internal/utils"
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
		"version of the IP protocol, supports 'v4' | 'v6' | 'all'",
	)

	configFlags["config"] = flag.String(
		"config",
		"config.yml",
		"path to the configuration file",
	)

	configFlags["exec"] = flag.String(
		"exec",
		"on_change",
		"run executable/script upon an event, supports 'on_change' | 'on_same' | 'always' | 'never'",
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

	config.Init(configFlags)

	database.ConnectDatabase()
	db := database.GetDatabase()

	err := db.AutoMigrate(&database.AddressEntry{})
	utils.Check(err, "could not run AutoMigrate")

}
