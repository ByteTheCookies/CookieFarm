// Package config for configuration management
package config

import (
	"github.com/ByteTheCookies/CookieFarm/pkg/models"
	"github.com/ByteTheCookies/CookieFarm/pkg/protocols"
)

var SharedConfig models.ConfigShared // Initialize the config struct

var (
	Debug         bool                                                                 // Global debug flag
	UseConfigFile bool                                                                 // Use configuration file instead of web config
	Password      string                                                               // Password for authentication
	ServerPort    string                                                               // Port for server
	Secret        = make([]byte, 32)                                                   // JWT secret key
	Submit        func(string, string, []string) ([]protocols.ResponseProtocol, error) // Function to submit data
	Cache         = true                                                               // Cache static file like css/js/image (If cache is enable more ram is used [default:true])
	NATSToken     string                                                               // NATS token for authentication
)

const (
	ConfigPath    string = "config.yml" // Path to configuration file
	DefaultLimit  int    = 100          // Default maximum number of flags to retrieve in the view
	DefaultOffset int    = 0            // Default offset for pagination
)
