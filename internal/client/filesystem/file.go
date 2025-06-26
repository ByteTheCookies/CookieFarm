package filesystem

import (
	"os"
	"path/filepath"
	"regexp"
)

// pathRegex is a regular expression to match paths starting with ~.
var pathRegex = regexp.MustCompile(`(~)([^/]*)(/?.*)`)

// GetExecutableDir returns the directory of the executable.
func GetExecutableDir() string {
	exePath, err := os.Executable()
	if err != nil {
		panic("impossible to determine the binary path: " + err.Error())
	}
	return filepath.Dir(exePath)
}

// IsValidFile checks if a file exists and is writable. If it does not exist, it attempts to create it.
func IsValidFile(fp string) bool {
	if _, err := os.Stat(fp); err == nil {
		return true
	}

	var d []byte
	if err := os.WriteFile(fp, d, 0o644); err == nil {
		os.Remove(fp)
		return true
	}

	return false
}
