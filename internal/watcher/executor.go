package watcher

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gweebg/ipwatcher/internal/config"
	"github.com/rs/zerolog"
)

// Executor is used to execute actions when an event is triggered.
// Needs an error channel to be passed, to be able to indicate when
// errors occur while executing the actions.
type Executor struct {
	Timeout   time.Duration
	logger    zerolog.Logger
	errorChan chan error
}

// NewExecutor creates a config.Exec executor.
//
// Usage of a watcher.Executor
//
//		   ex := NewExecutor(errorChannel)
//		   action = config.Exec{
//		       Type: "python",
//	        Args: "",
//			   Path: "script.py",
//		   }
//		   ex.Execute(action)
//
// Each execution is associated with a context.ContextWithTimeout delimiting
// the maximum time the action has to execute, defined on the configuration file.
func NewExecutor(errorChan chan error) *Executor {

	c := config.GetConfig()

	timeout := c.GetInt64("watcher.default_ttl")
	if timeout == 0 {
		timeout = 60
	}

	return &Executor{
		logger:    GetLogger().With().Str("service", "executor").Logger(),
		Timeout:   time.Duration(timeout) * time.Second,
		errorChan: errorChan,
	}
}

// ExecuteSlice executes, in parallel, a slice of config.Exec actions.
func (e *Executor) ExecuteSlice(actions []config.ExecuteAction) {
	for _, action := range actions {
		e.logger.Debug().Str("command", action.String()).Msg("executing action")
		go e.Execute(action)
	}
}

// Execute executes the given config.Exec action defined on the configuration
// file under 'events.<event>.actions'. Runs the action with a timed out context.Context
// killing the process if a configuration file defined threshold (in seconds) is crossed,
// limiting the execution time of the action.
func (e *Executor) Execute(action config.ExecuteAction) {

	cmd, ctx, cancel := action.Command(e.Timeout)
	if cancel != nil && ctx != nil {
		log.Println("with timeout!!!")
		defer cancel()

		// control the execution time of the current action
		go func() {
			<-ctx.Done()
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				e.logger.Warn().Str("action", action.String()).Err(fmt.Errorf("execution time exceeded (max is %v)", e.Timeout)).Send()
				_ = cmd.Process.Kill() // try to kill just in case of children processes
			}
		}()
	}

	stdout, _ := cmd.StdoutPipe()
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			e.logger.Debug().Str("command", action.String()).Msg(scanner.Text())
		}
	}()

	// redirecting the stderr of the spawned process to the pipe for later logging
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		e.errorChan <- errors.Join(err, ErrorExecutor)
		return
	}

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		e.errorChan <- errors.Join(errors.New(scanner.Text()), ErrorExecutor)
	}

	err := cmd.Wait()
	if err != nil {
		e.errorChan <- errors.Join(err, ErrorExecutor)
		return
	}

	e.logger.Debug().Str("command", action.String()).Msg("finished executing")
}
