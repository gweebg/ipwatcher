package watcher

import (
	"errors"
	"github.com/gweebg/ipwatcher/internal/config"
	"github.com/gweebg/ipwatcher/internal/database"
	"github.com/gweebg/ipwatcher/internal/utils"
	"log"
	"time"
)

type Watcher struct {

	// Version indicates the versions the watcher is supposed to track (v4|v6|all)
	Version string

	// allowApi exposes an API with the database records
	allowApi bool
	// allowNotif enables notification via email
	allowNotif bool
	// allowExec enables the execution of actions upon an event
	allowExec bool

	// Timeout represents the duration between each address query
	Timeout time.Duration

	// ticker is a *time.Ticker object responsible for waiting Timeout
	ticker *time.Ticker
	// tickerQuitChan allows the stop of the ticker
	tickerQuitChan chan struct{}
}

func NewWatcher() *Watcher {

	c := config.GetConfig()

	timeout := time.Duration(
		c.GetInt("watcher.timeout")) * time.Second

	return &Watcher{
		Version: c.GetString("flags.version"),

		allowApi:   c.GetBool("flags.api"),
		allowNotif: c.GetBool("flags.notify"),
		allowExec:  c.GetBool("flags.exec"),

		Timeout:        timeout,
		ticker:         time.NewTicker(timeout),
		tickerQuitChan: make(chan struct{}),
	}
}

func (w *Watcher) HandleEvent(eventType string) error {

	c := config.GetConfig()

	var handler *config.EventHandler
	events := c.Get("watcher.events").(*config.Events)

	switch eventType {

	case "OnChange":
		handler = events.OnChange
	case "OnMatch":
		handler = events.OnMatch
	case "OnError":
		handler = events.OnError

	default:
		return errors.New("unknown event type: " + eventType)
	}

	if handler != nil && w.allowExec {

		for _, exec := range handler.Actions {
			if handler.Notify && w.allowNotif {
				log.Printf("sent notification to destination\n")
			}
			log.Printf("executing '%s %s %s'\n\n", exec.Type, exec.Path, exec.Args)
		}
	}

	return nil
}

func (w *Watcher) Watch() {
	// todo : stop upon ctrl-c

	var records = new(database.AddressEntry)

	go func() {

		for {
			select {

			case <-w.ticker.C:

				// get the address from the desired source
				address, err := RequestAddress(w.Version)
				if err != nil {
					log.Printf("weh weh weh")
				}

				// get latest address record of the database
				previousAddress, err := records.First(w.Version)

				// if the database is empty, then we insert the current address
				if previousAddress == nil {
					_, err = records.Create(address, w.Version, address)
					continue
				}

				// compare addresses and handle accordingly
				if address != previousAddress.Address {

					_, err = records.Create(address, w.Version, previousAddress.Address) // insert new record onto the database
					utils.Check(err, "")

					err = w.HandleEvent("OnChange")

				} else {
					err = w.HandleEvent("OnMatch")
				}

				if err != nil {
					err = w.HandleEvent("OnError")
				}

			case <-w.tickerQuitChan:
				w.ticker.Stop()

			}
		}

	}()
}
