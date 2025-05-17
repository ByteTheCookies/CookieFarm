// Package utils provides utility functions for the CookieFarm client.
package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/ByteTheCookies/cookieclient/internal/config"
	"github.com/ByteTheCookies/cookieclient/internal/models"
)

const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Gray    = "\033[37m"
	White   = "\033[97m"
)

// Detach detaches the current process from the terminal re executing itself.
func Detach() {
	cmd := exec.Command(os.Args[0], os.Args[1:]...)

	filteredArgs := []string{}
	for _, arg := range os.Args[1:] {
		if arg != "--detach" && arg != "-d" {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	cmd = exec.Command(os.Args[0], filteredArgs...)

	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	err := cmd.Start()
	if err != nil {
		fmt.Println(Red+"[ERROR]"+Reset+" | Error during detach:", err)
		os.Exit(1)
	}

	fmt.Println(Yellow+"[WARN]"+Reset+" | Process detached with PID:", cmd.Process.Pid)
	os.Exit(0)
}

// MapPortToService maps a port to a service name.
func MapPortToService(port uint16) string {
	for _, service := range config.Current.ConfigClient.Services {
		if service.Port == port {
			return service.Name
		}
	}
	return ""
}

// GetExecutableDir returns the directory of the executable.
func GetExecutableDir() string {
	exePath, err := os.Executable()
	if err != nil {
		panic("impossible to determine the binary path: " + err.Error())
	}
	return filepath.Dir(exePath)
}

// ValidateArgs validates the arguments passed to the program.
func ValidateArgs(args models.Args) error {
	if *args.ExploitPath == "" {
		return errors.New("missing required --exploit argument")
	}

	if *config.HostServer == "" {
		return errors.New("missing required --base_url_server argument")
	}
	if *args.Password == "" {
		return errors.New("missing required --password argument")
	}

	if *args.TickTime < 1 {
		return errors.New("tick time must be at least 1")
	}

	exploitPath, err := filepath.Abs(*args.ExploitPath)
	if err != nil {
		return fmt.Errorf("error resolving exploit path: %v", err)
	}

	if info, err := os.Stat(exploitPath); err == nil && info.Mode()&0o111 == 0 {
		return fmt.Errorf("exploit file is not executable")
	}

	if _, err := os.Stat(exploitPath); os.IsNotExist(err) {
		return errors.New("exploit not found in the exploits directory")
	}

	return nil
}
