package filesystem

import (
	"os"
	"os/user"
	"path"
	"strings"
)

// IsPath checks if the provided string is a path exploit.
func IsPath(pathExploit string) bool {
	if strings.HasPrefix(pathExploit, "/") || strings.HasPrefix(pathExploit, ".") || strings.HasPrefix(pathExploit, "~") {
		return true
	}
	return false
}

// Code by @prep on Github https://github.com/prep/tilde
func ExpandTilde(p string) (string, error) {
	if len(p) < 1 || p[0] != '~' {
		return p, nil
	}

	var tildePath string

	results := pathRegex.FindStringSubmatch(p)[2:]
	switch results[0] {
	case "":
		u, err := user.Current()
		if err != nil {
			return "", err
		}

		tildePath = u.HomeDir
	case "+":
		pwd, err := os.Getwd()
		if err != nil {
			return "", err
		}

		tildePath = pwd
	default:
		u, err := user.Lookup(results[0])
		if err != nil {
			return "", err
		}

		tildePath = u.HomeDir
	}

	return path.Join(tildePath, results[1]), nil
}

// NormalizeNamePathExploit normalizes a given name path exploit by ensuring it has a .py extension
func NormalizeNamePathExploit(name string) (string, error) {
	if !strings.HasSuffix(name, ".py") {
		name += ".py"
	}

	var err error
	if strings.HasPrefix(name, "~") {
		name, err = ExpandTilde(name)
		if err != nil {
			return "", err
		}
	}

	return name, nil
}
