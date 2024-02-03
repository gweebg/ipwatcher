package watcher

import (
	"context"
	"errors"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
	"time"

	"github.com/gweebg/ipwatcher/internal/config"
	"github.com/gweebg/ipwatcher/internal/database"
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
	// fetcher is responsible for fetching information relative to the address
	fetcher *Fetcher
	// executor allows for the execution of configuration defined actions
	executor *Executor

	// Timeout represents the duration between each address query
	Timeout time.Duration
	// ticker is a *time.Ticker object responsible for waiting Timeout
	ticker *time.Ticker

	// tickerQuitChan allows the stop of the ticker
	tickerQuitChan chan struct{}
	// errorChan handles errors coming from the event handlers
	errorChan chan error
	// logger is the logger for this service
	logger zerolog.Logger
}

func NewWatcher() *Watcher {

	c := config.GetConfig()

	timeout := time.Duration(
		c.GetInt("watcher.timeout")) * time.Second

	var notifier *Notifier = nil
	if c.GetBool("flags.notify") {
		notifier = NewNotifier()
	}

	errorChan := make(chan error)

	var executor *Executor = nil
	if c.GetBool("flags.exec") {
		executor = NewExecutor(errorChan)
	}

	return &Watcher{
		Version:   c.GetString("flags.version"),
		allowApi:  c.GetBool("flags.api"),
		allowExec: c.GetBool("flags.exec"),

		notifier: notifier,
		fetcher:  NewFetcher(),
		executor: executor,

		Timeout: timeout,
		ticker:  time.NewTicker(timeout),

		tickerQuitChan: make(chan struct{}),
		errorChan:      errorChan,
		logger:         GetLogger().With().Str("service", "watcher").Logger(),
	}
}

func (w *Watcher) Watch() {

	w.logger.Info().Msg("watcher service is now running")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go w.check()

	for sig := range c {
		w.logger.Warn().Msgf("received %v signal, stopping watcher...", sig.String())
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
		w.logger.Fatal().Msgf("unknown event type '%v', skipping", eventType)
	}

	if handler != nil {

		if handler.Notify && w.notifier != nil {

			err := w.notifier.NotifyMail(ctx)
			if err != nil {
				w.errorChan <- errors.Join(err, ErrorNotifier)
			}
			w.logger.Info().
				Str("event", eventType).
				Msgf("recipients (%d) notified", len(w.notifier.Recipients))
		}

		if w.executor != nil {
			w.executor.ExecuteSlice(handler.Actions)
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

		w.logger.Error().Err(err).Msg("unexpected error")
	}
}

func (w *Watcher) check() {

	go w.errors()

	var records = new(database.AddressEntry)
	for {
		select {

		case <-w.ticker.C:

			// get the address from the desired source
			address, source, err := w.fetcher.RequestAddress(w.Version)
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

				w.logger.Info().
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

				w.logger.Info().Msgf("no address changes")

				ctx = context.WithValue(ctx, "source", source)
				go w.HandleEvent("on_match", ctx) // handle on_match
			}

		case <-w.tickerQuitChan:
			w.ticker.Stop()

		}
	}
}
