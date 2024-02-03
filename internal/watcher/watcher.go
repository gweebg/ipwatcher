package watcher

import (
	"context"
	"errors"
	"github.com/gweebg/ipwatcher/internal/config"
	"github.com/gweebg/ipwatcher/internal/database"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
	"time"
)

var logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
	Level(zerolog.TraceLevel).
	With().
	Timestamp().
	Caller().
	Logger()

var watcherLogger = logger.With().Str("service", "watcher").Logger()

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
		Timeout:  timeout,
		ticker:   time.NewTicker(timeout),

		tickerQuitChan: make(chan struct{}),
		errorChan:      make(chan error),
	}
}

func (w *Watcher) Watch() {

	watcherLogger.Info().Msg("watcher service is now running")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go w.check()

	for sig := range c {
		watcherLogger.Warn().Msgf("received %v signal, stopping watcher...", sig.String())
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
		watcherLogger.Fatal().Msgf("unknown event type '%v', skipping", eventType)
	}

	if handler != nil {

		if handler.Notify && w.notifier != nil {
			err := w.notifier.NotifyMail(ctx)
			if err != nil {
				w.errorChan <- errors.Join(err, ErrorNotifier)
			}
			watcherLogger.Debug().
				Str("event", eventType).
				Msgf("recipients (%d) notified", len(w.notifier.Recipients))
		}

		for _, exec := range handler.Actions {
			watcherLogger.Info().Msgf("executing '%s %s %s'\n\n", exec.Type, exec.Path, exec.Args)
		}
	}
}

func (w *Watcher) errors() {

	ctx := context.Background()
	ctx = context.WithValue(ctx, "timestamp", time.Now())

	for err := range w.errorChan {

		if !errors.Is(err, ErrorNotifier) {
			ctx = context.WithValue(ctx, "error", err)
			w.HandleEvent("on_error", ctx) // handle on_error
		}

		watcherLogger.Error().Err(err).Msg("unexpected error")
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

				watcherLogger.Info().
					Str("previous_address", previousAddress.Address).
					Str("current_address", address).
					Msgf("detected address change")

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

				watcherLogger.Debug().Msgf("no address changes")

				ctx = context.WithValue(ctx, "source", source)
				go w.HandleEvent("on_match", ctx) // handle on_match
			}

		case <-w.tickerQuitChan:
			w.ticker.Stop()

		}
	}
}
