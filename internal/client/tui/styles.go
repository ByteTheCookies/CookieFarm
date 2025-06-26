package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("#CDA157")
	secondaryColor = lipgloss.Color("#D9D9D9")
	textColor      = lipgloss.Color("#EDEDED")
	mutedColor     = lipgloss.Color("#888888")
	errorColor     = lipgloss.Color("#E74C3C")
	successColor   = lipgloss.Color("#219B54")
	warningColor   = lipgloss.Color("#FFC107")
	infoColor      = lipgloss.Color("#2196F3")

	// General styles
	AppStyle = lipgloss.NewStyle().
			Margin(1, 2)

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(textColor).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)

	// Banner style
	BannerStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1)

	// Input styles
	InputLabelStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	// Message styles
	ErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(errorColor).
			Padding(1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(successColor).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(successColor).
			Padding(1)

	WarningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(warningColor).
			Padding(1)

	InfoStyle = lipgloss.NewStyle().
			Foreground(infoColor).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(infoColor).
			Padding(1)

	// Help style
	HelpStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Menu styles
	MenuTitleStyle = lipgloss.NewStyle().
			Background(primaryColor).
			Foreground(textColor).
			Padding(0, 1)

	// Spinner styles
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	// Command output style
	CommandOutputStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(1)

	// Loading indicator style
	LoadingStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)
	// Process list styles
	ProcessListStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(1)

	ProcessHeaderStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true)

	ProcessSelectedStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	// Footer styles
	FooterStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	ListStyle = list.DefaultItemStyles{
		NormalTitle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#F7CD82")),
		NormalDesc:    lipgloss.NewStyle().Foreground(lipgloss.Color("#5B584F")),
		SelectedTitle: lipgloss.NewStyle().Foreground(lipgloss.Color("#F2F1EF")).Bold(true),
		SelectedDesc:  lipgloss.NewStyle().Foreground(lipgloss.Color("#E0D5AD")),
	}

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	TableBorderStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	TableRowStyle = lipgloss.NewStyle().
			Foreground(textColor)

	TableSelectedRowStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)
)

// DimmedStyle returns a dimmed version of the given text
func DimmedStyle(text string) string {
	return lipgloss.NewStyle().
		Foreground(mutedColor).
		Render(text)
}

// HighlightStyle returns a highlighted version of the given text
func HighlightStyle(text string) string {
	return lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Render(text)
}

// SuccessText returns a success-styled version of the given text
func SuccessText(text string) string {
	return lipgloss.NewStyle().
		Foreground(successColor).
		Bold(true).
		Render(text)
}

// ErrorText returns an error-styled version of the given text
func ErrorText(text string) string {
	return lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true).
		Render(text)
}

// WarningText returns a warning-styled version of the given text
func WarningText(text string) string {
	return lipgloss.NewStyle().
		Foreground(warningColor).
		Bold(true).
		Render(text)
}

// InfoText returns an info-styled version of the given text
func InfoText(text string) string {
	return lipgloss.NewStyle().
		Foreground(infoColor).
		Bold(true).
		Render(text)
}
