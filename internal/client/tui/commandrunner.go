package tui

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ByteTheCookies/CookieFarm/cmd/client/cmd"
	"github.com/ByteTheCookies/CookieFarm/internal/client/config"
	"github.com/ByteTheCookies/CookieFarm/internal/client/exploit"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"

	tea "github.com/charmbracelet/bubbletea"
)

// NewCommandRunner creates a new command runner
func NewCommandRunner() *CommandRunner {
	return &CommandRunner{
		exploitOutputChan: make(chan ExploitOutput, 100),
	}
}

// GetExploitOutputCmd returns a tea.Cmd that streams exploit output
func (r *CommandRunner) GetExploitOutputCmd() tea.Cmd {
	return func() tea.Msg {
		select {
		case output, ok := <-r.exploitOutputChan:
			if !ok {
				return nil
			}
			return output
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	}
}

// CloseExploitOutput closes the exploit output channel
func (r *CommandRunner) CloseExploitOutput() {
	if r.exploitOutputChan != nil {
		close(r.exploitOutputChan)
	}
}

// GetCurrentExploitPID returns the PID of the currently running exploit
func (r *CommandRunner) GetCurrentExploitPID() int {
	return r.currentExploitPID
}

// ExecuteCommand executes a generic shell command and returns its output
func (*CommandRunner) ExecuteCommand(command string, args ...string) (string, error) {
	cmdH := exec.Command(command, args...)

	var stdout, stderr bytes.Buffer
	cmdH.Stdout = &stdout
	cmdH.Stderr = &stderr

	err := cmdH.Run()

	var output strings.Builder

	if stdout.Len() > 0 {
		output.WriteString(stdout.String())
	}

	if stderr.Len() > 0 {
		if output.Len() > 0 {
			output.WriteString("\n\n")
		}
		output.WriteString("Error output:\n")
		output.WriteString(stderr.String())
	}

	if err != nil && output.Len() == 0 {
		return "", err
	}

	return output.String(), err
}

// ExecuteConfigCommand executes configuration-related commands
func (*CommandRunner) ExecuteConfigCommand(subcommand string) (string, error) {
	cm := config.GetConfigManager()
	switch subcommand {
	case "show":
		return cm.ShowLocalConfigContent()
	case "reset":
		return cm.ResetLocalConfigToDefaults()
	case "logout":
		return cm.Logout()
	default:
		return "", fmt.Errorf("unknown config subcommand: %s", subcommand)
	}
}

// ExecuteExploitCommand executes exploit-related commands
func (*CommandRunner) ExecuteExploitCommand(subcommand string) (string, error) {
	// switch subcommand {
	// default:
	// 	return "", nil
	// }
	return "", nil
}

// ExecuteLogin handles the login command
func (*CommandRunner) ExecuteLogin(password, host, username string, port uint16, https bool) (string, error) {
	cm := config.GetConfigManager()
	configuration := config.ConfigLocal{
		Host:     host,
		Username: username,
		Port:     port,
		HTTPS:    https,
	}
	_, err := cm.SetLocalConfig(configuration)
	if err != nil {
		return "", fmt.Errorf("error during update of config in the file: %s", err)
	}
	cmd.Password = password
	pathSession, err := cmd.LoginHandler(password)
	if err != nil {
		return "", fmt.Errorf("login failed: %w", err)
	}
	return "Login successfully session saved at " + pathSession, nil
}

// ExecuteConfigUpdate handles the config update command
func (*CommandRunner) ExecuteConfigUpdate(host, port, username string, useHTTPS bool) (string, error) {
	cm := config.GetConfigManager()
	configuration := config.ConfigLocal{
		Host:     host,
		Username: username,
		HTTPS:    useHTTPS,
	}

	if port != "" {
		portNum, err := strconv.Atoi(port)
		if err == nil {
			configuration.Port = uint16(portNum)
		}
	}

	path, err := cm.SetLocalConfig(configuration)
	if err != nil {
		return "", fmt.Errorf("error during update of config in the file: %s", err)
	}
	return "Configuration updated successfully. File saved at:" + path, nil
}

func (*CommandRunner) ExecuteExploitTest(
	exploitPath, servicePort string,
	tickTime, threadCount string, submitValue bool,
) (string, error) {
	tickTimeInt := 120  // default
	threadCountInt := 5 // default
	var err error

	if tickTime != "" {
		tickTimeInt, err = strconv.Atoi(tickTime)
		if err != nil {
			return "", fmt.Errorf("invalid tick time: %s", tickTime)
		}
	}

	if threadCount != "" {
		threadCountInt, err = strconv.Atoi(threadCount)
		if err != nil {
			return "", fmt.Errorf("invalid thread count: %s", threadCount)
		}
	}

	if servicePort == "" {
		return "", errors.New("service port is required")
	}

	err = exploit.Run(
		exploitPath,
		tickTimeInt,
		threadCountInt,
		servicePort,
		false,
		submitValue,
	)
	if err != nil {
		return "", err
	}

	var initialOutput strings.Builder
	initialOutput.WriteString(fmt.Sprintf("Running with %d threads, tick time %d seconds\n", threadCountInt, tickTimeInt))
	initialOutput.WriteString("Output streaming started. Live updates will appear below...\n")

	return initialOutput.String(), nil
}

func (*CommandRunner) ExecuteExploitRun(
	exploitPath, serviceName string,
	tickTime, threadCount string, submitValue bool,
) (string, error) {
	tickTimeInt := 120  // default
	threadCountInt := 5 // default
	var err error

	if tickTime != "" {
		tickTimeInt, err = strconv.Atoi(tickTime)
		if err != nil {
			return "", fmt.Errorf("invalid tick time: %s", tickTime)
		}
	}

	if threadCount != "" {
		threadCountInt, err = strconv.Atoi(threadCount)
		if err != nil {
			return "", fmt.Errorf("invalid thread count: %s", threadCount)
		}
	}

	if serviceName == "" {
		return "", errors.New("service port is required")
	}

	err = exploit.Run(
		exploitPath,
		tickTimeInt,
		threadCountInt,
		serviceName,
		false,
		submitValue,
	)
	if err != nil {
		return "", err
	}

	var initialOutput strings.Builder
	initialOutput.WriteString(fmt.Sprintf("Running with %d threads, tick time %d seconds\n", threadCountInt, tickTimeInt))
	initialOutput.WriteString("Output streaming started. Live updates will appear below...\n")

	return initialOutput.String(), nil
}

// ExecuteExploitCreate handles creating an exploit template
func (*CommandRunner) ExecuteExploitCreate(name string) (string, error) {
	return exploit.Create(name)
}

// ExecuteExploitRemove handles removing an exploit template
func (*CommandRunner) ExecuteExploitRemove(name string) (string, error) {
	return exploit.Remove(name)
}

// GetRunningExploits returns a list of running exploit processes
func (*CommandRunner) GetRunningExploits() ([]ExploitProcess, error) {
	cm := config.GetConfigManager()
	if err := cm.LoadLocalConfigFromFile(); err != nil {
		return nil, fmt.Errorf("error loading configuration: %w", err)
	}

	id := 1
	processes := make([]ExploitProcess, 0, len(cm.GetLocalConfig().Exploits))
	filtered := make([]config.Exploit, 0, len(cm.GetLocalConfig().Exploits))
	for _, exploitS := range cm.GetLocalConfig().Exploits {
		proc, err := os.FindProcess(exploitS.PID)
		if err != nil || proc == nil || proc.Signal(syscall.Signal(0)) != nil {
			logger.Log.Warn().Str("exploit", exploitS.Name).Msg("Exploit removed due to invalid or inactive process")
			continue
		}
		processes = append(processes, ExploitProcess{
			ID:   id,
			Name: filepath.Base(exploitS.Name),
			PID:  exploitS.PID,
		})
		id++
		filtered = append(filtered, exploitS)
	}

	localConfig := cm.GetLocalConfig()
	localConfig.Exploits = filtered

	if _, err := cm.SetLocalConfig(localConfig); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to write configuration after filtering exploits")
	}
	return processes, nil
}

// ExecuteExploitStop handles stopping a running exploit
func (*CommandRunner) ExecuteExploitStop(pid string) (string, error) {
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return "", fmt.Errorf("invalid process ID: %s", pid)
	}
	cm := config.GetConfigManager()
	cm.SetPID(pidInt)
	return exploit.Stop(pidInt)
}
