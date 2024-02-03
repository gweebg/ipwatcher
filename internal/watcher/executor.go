package watcher

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/gweebg/ipwatcher/internal/config"
	"github.com/rs/zerolog"
	"os/exec"
	"strings"
	"time"
)

type Executor struct {
	Timeout   time.Duration
	logger    zerolog.Logger
	errorChan chan error
}

func NewExecutor(errorChan chan error) *Executor {

	c := config.GetConfig()

	timeout := c.GetInt64("watcher.max_execution_time")
	if timeout == 0 {
		timeout = 60
	}

	return &Executor{
		logger:    GetLogger().With().Str("service", "executor").Logger(),
		Timeout:   time.Duration(timeout) * time.Second,
		errorChan: errorChan,
	}
}

func (e *Executor) ExecuteSlice(actions []config.Exec) {
	for _, action := range actions {
		e.logger.Debug().Str("command", action.String()).Msg("executing action")
		go e.Execute(action)
	}
}

func (e *Executor) Execute(action config.Exec) {

	ctx, cancel := context.WithTimeout(context.Background(), e.Timeout)
	defer cancel()

	go func() {
		<-ctx.Done()
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			e.logger.Warn().Str("action", action.String()).Err(errors.New(fmt.Sprintf("execution time exceeded (max is %v)", e.Timeout))).Send()
		}
	}()

	args := strings.Split(action.Args, " ")
	args = append([]string{action.Path}, args...)

	cmd := exec.CommandContext(ctx, action.Type, args...)

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
