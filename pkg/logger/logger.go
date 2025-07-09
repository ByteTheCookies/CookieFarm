// Package logger provides functions to manage the CookieFarm client logging.
package logger

import (
	_ "embed"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog"
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

var (
	LogLevel zerolog.Level  // LogLevel defines the log level for the logger.
	Log      zerolog.Logger // Log is the logger instance for the CookieFarm client.
	LogFile  *os.File       // logFile represents the log file for the CookieFarm client.
	useTUI   bool           // NoTUI indicates whether to disable the TUI mode for logging
)

//go:embed banner.txt
var banner string

const defaultLogPath = "/tmp/cookielogs"

// SetTUI sets the useTUI variable to enable or disable TUI mode for logging.
func SetTUI(value bool) {
	useTUI = value
}

func setLevel(level string) {
	parsedLevel, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		parsedLevel = zerolog.InfoLevel
	}
	LogLevel = parsedLevel
}

func setupLogFile() {
	_ = os.MkdirAll(defaultLogPath, 0o755)
	logPath := filepath.Join(defaultLogPath, "cookiefarm-"+strconv.Itoa(int(time.Now().UnixMilli()))) + ".log"

	var err error
	LogFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0o666)
	if err != nil {
		panic("cannot create log file: " + err.Error())
	}
}

// Setup configures the logger with the specified log level and returns the log file path.
func Setup(level string, file bool) string {
	var logPath string
	setLevel(level)

	if file {
		setupLogFile()
	}

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
		FormatLevel: func(i any) string {
			lvl := strings.ToLower(fmt.Sprintf("%s", i))
			switch lvl {
			case "debug":
				return Gray + "[DEBUG]" + Reset
			case "info":
				return Cyan + "[INFO]" + Reset
			case "warn":
				return Yellow + "[WARN]" + Reset
			case "error":
				return Red + "[ERROR]" + Reset
			case "fatal":
				return Magenta + "[FATAL]" + Reset
			default:
				return lvl
			}
		},
		FormatMessage: func(i any) string {
			return fmt.Sprintf("%s", i)
		},
		FormatFieldName: func(i any) string {
			return White + fmt.Sprintf("%s=", i)
		},
		FormatFieldValue: func(i any) string {
			return fmt.Sprintf("%v", i) + Reset
		},
	}

	var multi zerolog.LevelWriter
	switch {
	case useTUI:
		multi = zerolog.MultiLevelWriter(LogFile)
	case file:
		multi = zerolog.MultiLevelWriter(consoleWriter, LogFile)
	default:
		multi = zerolog.MultiLevelWriter(consoleWriter)
	}

	if level == "debug" {
		Log = zerolog.New(multi).With().Timestamp().Caller().Logger()
	} else {
		Log = zerolog.New(multi).With().Timestamp().Logger()
	}

	zerolog.SetGlobalLevel(LogLevel)

	return logPath
}

// Close shuts down the logger by closing the log file if it is open.
func Close() {
	if LogFile != nil {
		_ = LogFile.Close()
	}
}

// GetBanner returns a formatted banner string with the specified data.
func GetBanner(data string) string {
	bannerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CDA157")).
		Bold(true).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#CDA157")).
		Padding(1, 2).
		MarginTop(1)
	formattedBanner := strings.ReplaceAll(banner, "<type>", data)
	return bannerStyle.Render(formattedBanner)
}

func PrintBanner(useBanner bool, data string) {
	if useBanner {
		fmt.Println(GetBanner(data))
	}
}

// IsCompletionCommand checks if the command line arguments indicate a completion command.
func IsCompletionCommand() bool {
	for _, arg := range os.Args {
		if strings.Contains(arg, "__complete") || strings.Contains(arg, "completion") {
			return true
		}
	}
	return false
}

var CookieCLIColorSchema = fang.ColorScheme{
	Base:           color.RGBA{0xe9, 0xe9, 0xe9, 0xe9},
	Title:          color.RGBA{0xCD, 0xA1, 0x57, 0xff},
	Description:    color.RGBA{0xD9, 0xD9, 0xD9, 0xff},
	Codeblock:      color.RGBA{0x0a, 0x0c, 0x0d, 0xff},
	Program:        color.RGBA{0xCD, 0xA1, 0x57, 0xff},
	DimmedArgument: color.RGBA{0x88, 0x88, 0x88, 0xff},
	Comment:        color.RGBA{0x88, 0x88, 0x88, 0xff},
	Flag:           color.RGBA{0x21, 0x96, 0xF3, 0xff},
	FlagDefault:    color.RGBA{0xD9, 0xD9, 0xD9, 0xff},
	Command:        color.RGBA{0xCD, 0xA1, 0x57, 0xff},
	QuotedString:   color.RGBA{0x21, 0x9B, 0x54, 0xff},
	Argument:       color.RGBA{0xED, 0xED, 0xED, 0xff},
	Help:           color.RGBA{0x88, 0x88, 0x88, 0xff},
	Dash:           color.RGBA{0x88, 0x88, 0x88, 0xff},
	ErrorHeader:    [2]color.Color{color.RGBA{0xED, 0xED, 0xED, 0xff}, color.RGBA{0xE7, 0x4C, 0x3C, 0xff}},
	ErrorDetails:   color.RGBA{0xE7, 0x4C, 0x3C, 0xff},
}
