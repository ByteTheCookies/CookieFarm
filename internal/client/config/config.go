// Package config provides functions to manage the CookieFarm client configuration globally.
package config

import "github.com/ByteTheCookies/CookieFarm/pkg/system"

var DefaultConfigPath, _ = system.ExpandTilde("~/.config/cookiefarm")

// Global instance for backward compatibility
var globalConfigManager = GetInstance()

var ExploitTemplate = []byte(`#!/usr/bin/env python3
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
`)

var ConfigTemplate = []byte(`host: "localhost"
port: 8080
https: false
username: "cookieguest"
`)

// GetConfigManager returns the global ConfigManager instance
// Use this to access the new configuration management system
func GetConfigManager() *ConfigManager {
	return globalConfigManager
}

// NewConfigManager creates a new ConfigManager instance for dependency injection
// Use this when you need a fresh instance (e.g., for testing)
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		useBanner: true,
	}
}
