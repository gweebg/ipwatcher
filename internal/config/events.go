package config

import (
	"errors"
	"reflect"
	"strings"
)

type Exec struct {
	// Type of the script (python|bash|binary)
	Type string `mapstructure:"type"`
	// Path to the script (full path)
	Path string `mapstructure:"path"`
	// Args sets the arguments for the script
	Args string `mapstructure:"args"`
}

func (s Exec) Validate() error {

	script := strings.ToLower(s.Type)
	if script != "python" && script != "bash" && script != "binary" {
		return errors.New("the 'type' field of an action can only be 'python', 'bash' or 'binary', not '" + script + "'")
	}

	if s.Path == "" {
		return errors.New("the 'path' field is mandatory when used on a action")
	}

	return nil
}

type EventHandler struct {
	// Notify enables notifications via email upon an event
	Notify bool `mapstructure:"notify"`
	// Actions define the slice containing which actions to run upon an event
	Actions []Exec `mapstructure:"actions"`
}

func (e EventHandler) Validate() error {
	for _, exec := range e.Actions {
		err := exec.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

type Events struct {
	// OnChange event handler, information about what to do when the address updates
	OnChange *EventHandler `mapstructure:"on_change"` // ! nil if it does not exist
	// OnMatch event handler, information about what to do when the address matches the previous
	OnMatch *EventHandler `mapstructure:"on_match"`
	// OnError event handler, information about what to do when an error occurs
	OnError *EventHandler `mapstructure:"on_error"`
}

func getEvents() (*Events, error) {

	config := GetConfig()
	if config == nil {
		return nil, errors.New("the 'sources' field can only be acquired after config initialization")
	}

	var events Events

	err := config.UnmarshalKey("watcher.events", &events)
	if err != nil {
		return nil, err
	}

	err = validateEvents(events)
	if err != nil {
		return nil, err
	}

	return &events, nil
}

func validateEvents(events Events) error {

	v := reflect.ValueOf(events)

	for i := 0; i < v.NumField(); i++ {

		value := v.Field(i).Interface().(*EventHandler)
		if value != nil {

			err := value.Validate()
			if err != nil {
				return err
			}

		}
	}

	return nil
}
