// Package config provides functions to manage the CookieFarm client configuration globally.
package config

import "github.com/ByteTheCookies/cookieclient/internal/models"

var (
	Current    models.Config // Current holds the global configuration for the client.
	Token      string        // Token stores the global authentication token.
	HostServer *string       // BaseURLServer holds the global base URL for the server.
)
