//go:build ignore

package main

import (
	"bufio"
	"fmt"
	"models"
	"net"
	"protocols"
	"strings"
	"time"
)

func Submit(url string, teamToken string, flags []string) ([]protocols.ResponseProtocol, error) {
	_ = teamToken // plaintext TCP protocol does not use team token header

	conn, err := net.DialTimeout("tcp", tcpAddress(url), 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("error connecting to submission server: %w", err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		return nil, fmt.Errorf("error setting connection deadline: %w", err)
	}

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	readWelcomeBanner(reader)

	if err := writeFlags(writer, flags); err != nil {
		return nil, err
	}

	return readResponses(reader, flags)
}

func tcpAddress(url string) string {
	address := strings.TrimPrefix(url, "tcp://")
	address = strings.TrimPrefix(address, "tcp:")
	return strings.TrimPrefix(address, "//")
}

func readWelcomeBanner(reader *bufio.Reader) {
	// Optional welcome banner terminated by an empty line.
	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\n" {
			return
		}
	}
}

func writeFlags(writer *bufio.Writer, flags []string) error {
	for _, flag := range flags {
		if _, err := writer.WriteString(flag + "\n"); err != nil {
			return fmt.Errorf("error writing flag: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing flags: %w", err)
	}
	return nil
}

func readResponses(reader *bufio.Reader, flags []string) ([]protocols.ResponseProtocol, error) {
	responsesParsed := make([]protocols.ResponseProtocol, len(flags))
	flagIndex := buildFlagIndex(flags)

	for i := 0; i < len(flags); i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading response: %w", err)
		}

		if err := parseResponseLine(strings.TrimSuffix(line, "\n"), flagIndex, responsesParsed); err != nil {
			return nil, err
		}
	}

	return responsesParsed, nil
}

func buildFlagIndex(flags []string) map[string][]int {
	flagIndex := make(map[string][]int, len(flags))
	for i, f := range flags {
		flagIndex[f] = append(flagIndex[f], i)
	}
	return flagIndex
}

func parseResponseLine(line string, flagIndex map[string][]int, responsesParsed []protocols.ResponseProtocol) error {
	matchedFlag := matchFlag(line, flagIndex)
	if matchedFlag == "" {
		return fmt.Errorf("malformed response line: %q", line)
	}

	idx, err := nextFlagIndex(matchedFlag, flagIndex)
	if err != nil {
		return err
	}

	code, msg, err := parseStatusAndMessage(line, matchedFlag)
	if err != nil {
		return err
	}

	responsesParsed[idx].Flag = matchedFlag
	responsesParsed[idx].Msg = msg
	responsesParsed[idx].Status = statusFromCode(code)
	return nil
}

func matchFlag(line string, flagIndex map[string][]int) string {
	for f := range flagIndex {
		if hasFlagPrefix(line, f) {
			return f
		}
	}
	return ""
}

func hasFlagPrefix(line string, flag string) bool {
	rest, ok := strings.CutPrefix(line, flag)
	return ok && (rest == "" || rest[0] == ' ' || rest[0] == '\t')
}

func nextFlagIndex(matchedFlag string, flagIndex map[string][]int) (int, error) {
	indices := flagIndex[matchedFlag]
	if len(indices) == 0 {
		return 0, fmt.Errorf("unexpected duplicate response for flag %q", matchedFlag)
	}

	idx := indices[0]
	flagIndex[matchedFlag] = indices[1:]
	return idx, nil
}

func parseStatusAndMessage(line string, matchedFlag string) (string, string, error) {
	rest := strings.TrimLeft(line[len(matchedFlag):], " \t")
	parts := strings.Fields(rest)
	if len(parts) == 0 {
		return "", "", fmt.Errorf("missing status code in response: %q", line)
	}

	return parts[0], strings.Join(parts[1:], " "), nil
}

func statusFromCode(code string) models.Status {
	switch code {
	case "OK":
		return models.StatusAccepted
	case "DUP", "OWN", "OLD":
		return models.StatusDenied
	case "ERR":
		return models.StatusError
	case "INV":
		return models.StatusNotValid
	default:
		return models.StatusError
	}
}
