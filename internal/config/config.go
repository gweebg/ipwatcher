package config

import (
	"github.com/gweebg/ipwatcher/internal/utils"
	"github.com/spf13/viper"
	"reflect"
)

var config *viper.Viper

func Init(userFlags map[string]interface{}) {

	// todo: hot-reload upon configuration changes, use viper

	config = viper.New()

	config.SetConfigType("yaml")
	config.SetConfigName("config")
	config.AddConfigPath("./")
	config.AddConfigPath("config/")

	err := config.ReadInConfig()
	utils.Check(err, "")

	for key, val := range userFlags {
		rv := reflect.ValueOf(val)
		config.Set("flags."+key, rv.Elem().Interface())
	} // append the flags set by the user to the configuration object

	parsedSources, err := GetSources()
	utils.Check(err, "")
	config.Set("sources", parsedSources)

	parsedEvents, err := GetEvents()
	utils.Check(err, "")
	config.Set("watcher.events", parsedEvents)

}

func GetConfig() *viper.Viper {
	return config
}
