package watcher

import (
	"context"
	"errors"
	"github.com/gweebg/ipwatcher/internal/config"
	"github.com/gweebg/ipwatcher/internal/database"
	"log"
	"os"
	"os/signal"
	"time"
)

var (
	ErrorDatabase = errors.New("database error")
	ErrorNotifier = errors.New("notifier error")
	ErrorExecutor = errors.New("executor error")
	ErrorFetch    = errors.New("fetch error")
)

type Watcher struct {

	// Version indicates the versions the watcher is supposed to track (v4|v6|all)
	Version string

	// allowApi exposes an API with the database records
	allowApi bool
	// allowExec enables the execution of actions upon an event
	allowExec bool

	// notifier allows for email notification sending
	notifier *Notifier

	// Timeout represents the duration between each address query
	Timeout time.Duration
	// ticker is a *time.Ticker object responsible for waiting Timeout
	ticker *time.Ticker
	// tickerQuitChan allows the stop of the ticker
	tickerQuitChan chan struct{}
	// errorChan handles errors coming from the event handlers
	errorChan chan error
}

func NewWatcher() *Watcher {

	c := config.GetConfig()

	timeout := time.Duration(
		c.GetInt("watcher.timeout")) * time.Second

	var notifier *Notifier = nil
	if c.GetBool("flags.notify") {
		notifier = NewNotifier()
	}

	return &Watcher{
		Version:   c.GetString("flags.version"),
		allowApi:  c.GetBool("flags.api"),
		allowExec: c.GetBool("flags.exec"),

		notifier: notifier,

		Timeout:        timeout,
		ticker:         time.NewTicker(timeout),
		tickerQuitChan: make(chan struct{}),
		errorChan:      make(chan error),
	}
}

func (w *Watcher) Watch() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go w.check()

	for sig := range c {
		log.Printf("received %v signal, stopping the application\n", sig.String())
		w.Stop()
		return
	}
}

func (w *Watcher) Stop() {
	close(w.tickerQuitChan)
	close(w.errorChan)
}

func (w *Watcher) HandleEvent(eventType string, ctx context.Context) {

	c := config.GetConfig()
	ctx = context.WithValue(ctx, "event", eventType)

	var handler *config.EventHandler
	events := c.Get("watcher.events").(*config.Events)

	switch eventType {

	case "on_change":
		handler = events.OnChange
	case "on_match":
		handler = events.OnMatch
	case "on_error":
		handler = events.OnError

	default:
		log.Fatalf("unknown event type '%v', skipping\n", eventType)
	}

	if handler != nil {

		if handler.Notify && w.notifier != nil {
			err := w.notifier.NotifyMail(ctx)
			if err != nil {
				w.errorChan <- errors.Join(err, ErrorNotifier)
			}
		}

		for _, exec := range handler.Actions {
			log.Printf("executing '%s %s %s'\n\n", exec.Type, exec.Path, exec.Args)
		}
	}
}

func (w *Watcher) errors() {

	ctx := context.Background()

	for err := range w.errorChan {
		details := errors.Unwrap(err)

		if !errors.Is(err, ErrorNotifier) {
			ctx = context.WithValue(ctx, "error", details)
			w.HandleEvent("on_error", ctx) // handle on_error
		}

		log.Printf("an error occured:\n'%v'\n", err.Error())
	}
}

func (w *Watcher) check() {

	go w.errors()

	var records = new(database.AddressEntry)
	for {
		select {

		case <-w.ticker.C:

			// get the address from the desired source
			address, source, err := RequestAddress(w.Version)
			if err != nil {
				w.errorChan <- errors.Join(err, ErrorFetch)
				continue
			}

			// get latest address record of the database
			previousAddress, err := records.First(w.Version)
			if err != nil {
				w.errorChan <- errors.Join(err, ErrorDatabase)
				continue
			}

			// if the database is empty, then we insert the current address
			if previousAddress == nil {
				_, err = records.Create(address, w.Version, address)
				if err != nil {
					w.errorChan <- errors.Join(err, ErrorDatabase)
				}
				continue
			}

			ctx := context.Background()
			ctx = context.WithValue(ctx, "timestamp", time.Now())

			// compare addresses and handle accordingly
			if address != previousAddress.Address {

				_, err = records.Create(address, w.Version, previousAddress.Address) // insert new record onto the database
				if err != nil {
					w.errorChan <- errors.Join(err, ErrorDatabase)
					continue
				}

				ctx = context.WithValue(ctx, "previous_address", previousAddress.Address)
				ctx = context.WithValue(ctx, "current_address", address)
				ctx = context.WithValue(ctx, "source", source)

				go w.HandleEvent("on_change", ctx) // handle on_change

			} else {
				ctx = context.WithValue(ctx, "source", source)
				go w.HandleEvent("on_match", ctx) // handle on_match
			}

		case <-w.tickerQuitChan:
			w.ticker.Stop()

		}
	}
}
