package watcher

import (
	"github.com/gweebg/ipwatcher/internal/config"
	"time"
)

type Action func()
type EventHandlers struct {
	OnChange Action
	OnMatch  Action
	OnError  Action
}

func (e EventHandlers) Load() *EventHandlers {
	return nil
}

type Watcher struct {

	// Version indicates the versions the watcher is supposed to track (v4|v6|all)
	Version string

	// AllowNotification enables notifications via email upon an event
	AllowNotification bool
	// AllowApi exposes a API with the database records
	AllowApi bool
	// AllowExec enables the execution of the actions defined in the configuration file
	AllowExec bool

	// Timeout represents the duration between each address query
	Timeout time.Duration

	// ticker is a *time.Ticker object responsible for waiting Timeout
	ticker *time.Ticker
	// tickerQuitChan allows the stop of the ticker
	tickerQuitChan chan struct{}

	// Handlers contains the event handles for determined actions
	Handlers *EventHandlers
}

func NewWatcher() *Watcher {

	c := config.GetConfig()

	timeout := time.Duration(
		c.GetInt("watcher.timeout")) * time.Second

	allowExec := c.GetBool("flags.exec")
	allowNotif := c.GetBool("flags.notify")

	return &Watcher{
		Version: c.GetString("flags.version"),

		AllowNotification: allowNotif,
		AllowApi:          c.GetBool("flags.api"),
		AllowExec:         allowExec,

		Timeout:        timeout,
		ticker:         time.NewTicker(timeout),
		tickerQuitChan: make(chan struct{}),

		Handlers: nil,
	}
}

func (w *Watcher) Watch() error {
	go w.Loop()
	return nil
}

func (w *Watcher) Loop() {

	for {
		select {

		case <-w.ticker.C:
			// ! Request address
			// ! Get previous address
			// ! Compare addresses
			// ! Execute actions

		case <-w.tickerQuitChan:
			w.ticker.Stop()

		}
	}

}
