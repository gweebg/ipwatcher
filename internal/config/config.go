package config

import (
	"errors"
	"github.com/gweebg/ipwatcher/internal/utils"
	"github.com/spf13/viper"
	"reflect"
	"strings"
)

var config *viper.Viper

type SourceUrl struct {
	V4 *string `mapstructure:"v4"`
	V6 *string `mapstructure:"v6"`
}

func (s SourceUrl) GetUrl(version string) (string, error) {

	if strings.ToLower(version) == "v4" && s.V4 != nil {
		return *s.V4, nil
	}

	if strings.ToLower(version) == "v6" && s.V6 != nil {
		return *s.V6, nil
	}

	return "", errors.New("'" + version + "' is not specified in SourceUrl")

}

type Source struct {
	Name  string    `mapstructure:"name"`
	Url   SourceUrl `mapstructure:"url"`
	Type  string    `mapstructure:"type"`
	Field *string   `mapstructure:"field"`
}

func Init(userFlags map[string]interface{}) {
	// Todo: Maybe hot-reload ?
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
	}

	err = validateSources()
	utils.Check(err, "")

}

func GetSources() ([]Source, error) {

	var sources []Source

	if config == nil {
		return nil, errors.New("the 'sources' field can only be acquired after config initialization")
	}

	err := config.UnmarshalKey("sources", &sources)
	if err != nil {
		return nil, err
	}

	return sources, nil
}

func AddSource(newSource Source) error {

	sources := config.Get("sources").([]interface{})
	sources = append(sources, newSource)

	config.Set("sources", sources)
	if err := config.WriteConfig(); err != nil {
		return err
	}

	return nil
}

func GetConfig() *viper.Viper {
	return config
}

func validateSources() error {

	sources, err := GetSources()
	if err != nil {
		return err
	}

	for _, source := range sources {

		if !(source.Url.V4 != nil || source.Url.V6 != nil) {
			return errors.New("the 'url' field must have at 'v4' or 'v6' or both specified")
		}

		if source.Type == "json" && source.Field == nil {
			return errors.New("the 'field' field must be specified if 'response_type' is equal to 'json'")
		}

		if source.Type != "text" && source.Type != "json" {
			return errors.New("the field 'response_type' can only be 'text' or 'json'")
		}

	}

	return nil
}
