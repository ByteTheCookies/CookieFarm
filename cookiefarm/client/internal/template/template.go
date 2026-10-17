package template

import (
	"errors"
	"fmt"
	"logger"
	"os"
	"path/filepath"
	"system"

	"client/config"
)

const exploitTemplate string = `#!/usr/bin/env python3
from cookiefarm import exploit_manager

# "ip" are the IP address of the target team (example: 10.10.X.1)
# "port" is the port of the target service (example: 1337)
# "name_service" is the name of the service to exploit (example: "CookieService")
# "flag_ids" is the flag IDs of the target team and target service (example: [{"username": "psQSDAasd", "password": "qweqweqwe"}, {"username": "sdafjhAS", "password": "HIUOasdb"}])

@exploit_manager
def exploit(ip, port, name_service, flag_ids: list):
    # Run your exploit here
    flag = ""

    # Just print the flag to stdout
    print(flag)
`

func verifyAndHandlePath(path string) error {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		logger.Log.Warn().Msg("Default exploit path not exists... Creating it")

		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return fmt.Errorf("error creating exploit path: %v", err)
		}
	}

	return nil
}

func Create(name string) (string, error) {
	if name == "" {
		return "", errors.New("exploit name cannot be empty")
	}

	name, err := system.NormalizeNamePathExploit(name)
	if err != nil {
		return "", fmt.Errorf("error normalizing exploit name: %v", err)
	}

	var path string
	if system.IsPath(name) {
		path = name
		if err := verifyAndHandlePath(filepath.Dir(path)); err != nil {
			return "", err
		}
	} else {
		exploitPath := filepath.Join(config.DefaultPath, "exploits")
		if err := verifyAndHandlePath(exploitPath); err != nil {
			return "", err
		}
		path = filepath.Join(exploitPath, name)
	}
	exploitFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0o777)
	if err != nil {
		return "", fmt.Errorf("error creating exploit file: %v", err)
	}

	defer exploitFile.Close()
	exploitFile.WriteString(exploitTemplate)

	return "Exploit file created successfully at " + path, nil
}

func Remove(name string) (string, error) {
	if err := verifyAndHandlePath(config.DefaultPath); err != nil {
		return "", err
	}

	logger.Log.Debug().Str("Exploit name", name).Msg("Removing exploit template")

	namePathNormalized, err := system.NormalizeNamePathExploit(name)
	if err != nil {
		return "", fmt.Errorf("error normalizing exploit name: %v", err)
	}

	var path string

	if system.IsPath(namePathNormalized) {
		path = namePathNormalized
	} else {
		exploitsDir := filepath.Join(path, "exploits")
		if _, err := os.Stat(exploitsDir); os.IsNotExist(err) {
			logger.Log.Warn().Msg("Exploits directory does not exist, creating it")
			os.Mkdir(exploitsDir, os.ModePerm)
		}
		path = filepath.Join(config.DefaultPath, exploitsDir, namePathNormalized)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("exploit file does not exist: %s", path)
	}

	err = os.Remove(path)
	if err != nil {
		return "", fmt.Errorf("error removing exploit file: %v", err)
	}

	return "Exploit file removed successfully: " + path, nil
}
