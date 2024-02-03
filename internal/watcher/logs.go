package watcher

import (
	"github.com/gweebg/ipwatcher/internal/config"
	"github.com/rs/zerolog"
	"os"
	"time"
)

var logLevel zerolog.Level = zerolog.TraceLevel
var logWriter zerolog.LevelWriter

func InitLogger() {

	c := config.GetConfig()

	if c.GetBool("flags.quiet") {
		logLevel = zerolog.InfoLevel
	}

	// todo: check for flags.log_file, if exists enable logging to file as well
}

func GetLogger() zerolog.Logger {
	logWriter = zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	var logger = zerolog.New(logWriter).
		Level(logLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	return logger
}
