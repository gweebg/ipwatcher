package config

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"time"
)

type ExecuteAction struct {
	Type string `mapstructure:"type"`
	Bin  string `mapstructure:"bin"`
	Args string `mapstructure:"args"`
	TTL  int    `mapstructure:"ttl"`
}

func (s ExecuteAction) Command(ttl time.Duration) (*exec.Cmd, context.Context, context.CancelFunc) {

	if s.TTL < 0 {
		return exec.Command(s.Bin, strings.Split(s.Args, " ")...), nil, nil
	}

	if s.TTL > 0 { // ttl > 0, use this
		ttl = time.Duration(s.TTL) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	cmd := exec.CommandContext(ctx, s.Bin, strings.Split(s.Args, " ")...)
	return cmd, ctx, cancel
}

func (s ExecuteAction) Validate() error {

	if strings.ToLower(strings.TrimSpace(s.Type)) != "execute" {
		return errors.New("as for now, 'execute' is the only action type possible")
	}
	if !s.CheckInstalled() {
		return errors.New(
			"could not find executable for '" + s.Bin + "', you can check if it's installed by running 'which " + s.Bin + "'",
		)
	}
	return nil
}

func (s ExecuteAction) CheckInstalled() bool {
	_, err := exec.LookPath(s.Bin)
	return err == nil
}

func (s ExecuteAction) String() string {
	str := fmt.Sprintf("%v %v", s.Bin, s.Args)
	return strings.TrimSpace(str)
}

type EventHandler struct {
	Notify  bool            `mapstructure:"notify"`
	Actions []ExecuteAction `mapstructure:"actions"`
}

func (e EventHandler) Validate() error {
	for _, e := range e.Actions {
		err := e.Validate()
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
