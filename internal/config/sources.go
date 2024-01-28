package config

import (
	"errors"
	"strings"
)

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

func getSources() ([]Source, error) {

	var sources []Source

	if config == nil {
		return nil, errors.New("the 'sources' field can only be acquired after config initialization")
	}

	err := config.UnmarshalKey("sources", &sources)
	if err != nil {
		return nil, err
	}

	err = validateSources(sources)
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

func validateSources(sources []Source) error {

	for _, source := range sources {

		if !(source.Url.V4 != nil || source.Url.V6 != nil) {
			return errors.New("the 'url' field must have at 'v4' or 'v6' or both specified")
		}

		sourceType := strings.ToLower(source.Type)
		if sourceType == "json" && source.Field == nil {
			return errors.New("the 'field' field must be specified if 'response_type' is equal to 'json'")
		}

		if sourceType != "text" && sourceType != "json" {
			return errors.New("the field 'response_type' can only be 'text' or 'json'")
		}

	}

	return nil
}
