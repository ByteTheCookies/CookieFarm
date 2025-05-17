// Package executor provides functions to execute exploits and manage their output.
package executor

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strconv"

	"github.com/ByteTheCookies/cookieclient/internal/config"
	"github.com/ByteTheCookies/cookieclient/internal/flagparser"
	"github.com/ByteTheCookies/cookieclient/internal/logger"
	"github.com/ByteTheCookies/cookieclient/internal/models"
)

type ExecutionResult struct {
	Cmd       *exec.Cmd
	FlagsChan chan models.Flag
}

// Start starts the exploit_manager and listens for flags in stdout.
func Start(exploitPath, password string, tickTime int, threadCount int, logPath string) (*ExecutionResult, error) {
	cmd := exec.Command(
		exploitPath,
		*config.HostServer,
		password,
		strconv.Itoa(tickTime),
		config.Current.ConfigClient.RegexFlag,
		strconv.Itoa(threadCount),
		logPath,
	)

	logger.Log.Debug().
		Str("full path exploit", exploitPath).
		Int("tick time", tickTime).
		Str("command executed", cmd.String())

	logger.Log.Info().Msg("Starting exploiting process with exploit_manager...")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	flagsChan := make(chan models.Flag, 500)

	go readStdout(stdout, flagsChan)
	go readStderr(stderr)

	return &ExecutionResult{
		Cmd:       cmd,
		FlagsChan: flagsChan,
	}, nil
}

// LogParsedLineError logs an error based on the status and line.
func logParsedLineError(err error, status, line string) {
	switch status {
	case "fatal":
		logger.Log.Fatal().Err(err).Msg("Fatal error")
	case "info":
		logger.Log.Info().Err(err).Msg("Info")
	case "failed":
		logger.Log.Warn().Err(err).Msg("Failed to run exploit")
	default:
		logger.Log.Debug().Err(err).Msg("Parsing warning")
	}
	logger.Log.Debug().Str("raw", line).Msg("Raw line with error")
}

// Read the stdout and parse JSON lines into Flag structs.
func readStdout(stdout io.Reader, flagsChan chan<- models.Flag) {
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		flag, status, err := flagparser.ParseLine(line)
		if err != nil {
			logParsedLineError(err, status, line)
			continue
		}

		flagsChan <- flag
		logger.Log.Info().
			Str("flag", flag.FlagCode).
			Int("team", int(flag.TeamID)).
			Str("service", flag.ServiceName).
			Uint16("port", flag.PortService).
			Msg("Parsed and queued flag")
	}

	if err := scanner.Err(); err != nil {
		logger.Log.Error().Err(err).Msg("Error reading stdout scanner")
	}
}

// Read the stderr and log any errors.
func readStderr(stderr io.Reader) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		logger.Log.Warn().Str("stderr", scanner.Text()).Msg("Exploit stderr")
	}
	if err := scanner.Err(); err != nil {
		logger.Log.Error().Err(err).Msg("Error reading stderr scanner")
	}
}
