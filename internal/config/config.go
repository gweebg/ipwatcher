package config

import (
	"github.com/gweebg/ipwatcher/internal/utils"
	"github.com/spf13/viper"
	"reflect"
)

var (
	config *viper.Viper
)

func Init(userFlags map[string]interface{}) {

	config = viper.New()

	config.SetConfigType("yaml")
	config.SetConfigName("config")
	config.AddConfigPath("./")
	config.AddConfigPath("config/")

	load(userFlags)
}

func load(userFlags map[string]interface{}) {

	err := config.ReadInConfig()
	utils.Check(err, "")

	for key, val := range userFlags {
		rv := reflect.ValueOf(val)
		config.Set("flags."+key, rv.Elem().Interface())
	} // append the flags set by the user to the configuration object

	parsedSources, err := getSources()
	utils.Check(err, "")
	config.Set("sources", parsedSources)

	parsedEvents, err := getEvents()
	utils.Check(err, "")
	config.Set("watcher.events", parsedEvents)

}

func GetConfig() *viper.Viper {
	return config
}
