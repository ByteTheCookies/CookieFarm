package tui

import (
	"errors"
	"strconv"
	"strings"

	"github.com/ByteTheCookies/CookieFarm/internal/client/config"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ========== FormData creation function ==========

// CreateForm creates input fields for a specific command
func CreateForm(command string) ([]textinput.Model, []string) {
	var inputs []textinput.Model
	var labels []string

	switch command {
	case "config login":
		inputs, labels = createLoginForm()
	case "config update":
		inputs, labels = createConfigUpdateForm()
	case "exploit run", "exploit test":
		inputs, labels = createExploitRunForm()
	case "exploit create", "exploit remove":
		inputs, labels = createExploitNameForm()
	case "exploit stop":
		inputs, labels = createExploitStopForm()
	}

	if len(inputs) > 0 {
		inputs[0].Focus()
	}

	return inputs, labels
}

// createLoginForm creates form for login command
func createLoginForm() ([]textinput.Model, []string) {
	cm := config.GetConfigManager()
	var inputs []textinput.Model
	var labels []string

	passwordInput := textinput.New()
	passwordInput.Placeholder = "Password"
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.CharLimit = 32
	passwordInput.Width = 30
	inputs = append(inputs, passwordInput)
	labels = append(labels, "Password")

	// Host input
	hostInput := textinput.New()
	hostInput.Placeholder = "localhost"
	hostInput.CharLimit = 64
	hostInput.Width = 30
	hostInput.SetValue(cm.GetLocalConfig().Host)
	inputs = append(inputs, hostInput)
	labels = append(labels, "Host")

	// Port input
	portInput := textinput.New()
	portInput.Placeholder = "8080"
	portInput.CharLimit = 5
	portInput.Width = 10
	portInput.SetValue(strconv.Itoa(int(cm.GetLocalConfig().Port)))
	inputs = append(inputs, portInput)
	labels = append(labels, "Port")

	// Username input
	usernameInput := textinput.New()
	usernameInput.Placeholder = "cookieguest"
	usernameInput.CharLimit = 32
	usernameInput.Width = 30
	usernameInput.SetValue(cm.GetLocalConfig().Username)
	inputs = append(inputs, usernameInput)
	labels = append(labels, "Username")

	// HTTPS input
	httpsInput := textinput.New()
	httpsInput.Placeholder = "false"
	httpsInput.CharLimit = 5
	httpsInput.Width = 10
	if cm.GetLocalConfig().HTTPS {
		httpsInput.SetValue("true")
	} else {
		httpsInput.SetValue("false")
	}
	inputs = append(inputs, httpsInput)
	labels = append(labels, "HTTPS (true/false)")

	return inputs, labels
}

// createConfigUpdateForm creates form for config update command
func createConfigUpdateForm() ([]textinput.Model, []string) {
	cm := config.GetConfigManager()
	var inputs []textinput.Model
	var labels []string

	// Host input
	hostInput := textinput.New()
	hostInput.Placeholder = "localhost"
	hostInput.CharLimit = 64
	hostInput.Width = 30
	hostInput.SetValue(cm.GetLocalConfig().Host)
	inputs = append(inputs, hostInput)
	labels = append(labels, "Host")

	// Port input
	portInput := textinput.New()
	portInput.Placeholder = "8080"
	portInput.CharLimit = 5
	portInput.Width = 10
	portInput.SetValue(strconv.Itoa(int(cm.GetLocalConfig().Port)))
	inputs = append(inputs, portInput)
	labels = append(labels, "Port")

	// Username input
	usernameInput := textinput.New()
	usernameInput.Placeholder = "cookieguest"
	usernameInput.CharLimit = 32
	usernameInput.Width = 30
	usernameInput.SetValue(cm.GetLocalConfig().Username)
	inputs = append(inputs, usernameInput)
	labels = append(labels, "Username")

	// HTTPS input
	httpsInput := textinput.New()
	httpsInput.Placeholder = "false"
	httpsInput.CharLimit = 5
	httpsInput.Width = 10
	if cm.GetLocalConfig().HTTPS {
		httpsInput.SetValue("true")
	} else {
		httpsInput.SetValue("false")
	}
	inputs = append(inputs, httpsInput)
	labels = append(labels, "HTTPS (true/false)")

	return inputs, labels
}

// createExploitRunForm creates form for exploit run command
func createExploitRunForm() ([]textinput.Model, []string) {
	var inputs []textinput.Model
	var labels []string

	// Exploit path input
	exploitPathInput := textinput.New()
	exploitPathInput.Placeholder = "Path to exploit file"
	exploitPathInput.CharLimit = 256
	exploitPathInput.Width = 50
	inputs = append(inputs, exploitPathInput)
	labels = append(labels, "Exploit Path")

	// Service port input
	servicePortInput := textinput.New()
	servicePortInput.Placeholder = "Service port number"
	servicePortInput.CharLimit = 5
	servicePortInput.Width = 10
	inputs = append(inputs, servicePortInput)
	labels = append(labels, "Service Port")

	// Tick time input
	tickTimeInput := textinput.New()
	tickTimeInput.Placeholder = "120"
	tickTimeInput.CharLimit = 4
	tickTimeInput.Width = 10
	tickTimeInput.SetValue("120")
	inputs = append(inputs, tickTimeInput)
	labels = append(labels, "Tick Time (seconds)")

	// Thread count input
	threadCountInput := textinput.New()
	threadCountInput.Placeholder = "5"
	threadCountInput.CharLimit = 3
	threadCountInput.Width = 10
	threadCountInput.SetValue("5")
	inputs = append(inputs, threadCountInput)
	labels = append(labels, "Thread Count")

	// Submit value input
	submitValueInput := textinput.New()
	submitValueInput.Placeholder = "false"
	submitValueInput.CharLimit = 5
	submitValueInput.Width = 10
	submitValueInput.SetValue("false")
	inputs = append(inputs, submitValueInput)
	labels = append(labels, "Submit Value (true/false)")

	return inputs, labels
}

// createExploitNameForm creates form for exploit create/remove commands
func createExploitNameForm() ([]textinput.Model, []string) {
	var inputs []textinput.Model
	var labels []string

	// Exploit name input
	nameInput := textinput.New()
	nameInput.Placeholder = "Exploit name"
	nameInput.CharLimit = 64
	nameInput.Width = 40
	inputs = append(inputs, nameInput)
	labels = append(labels, "Exploit Name")

	return inputs, labels
}

// createExploitStopForm creates form for exploit stop command
func createExploitStopForm() ([]textinput.Model, []string) {
	var inputs []textinput.Model
	var labels []string

	pidInput := textinput.New()
	pidInput.Placeholder = "Select from list"
	pidInput.CharLimit = 10
	pidInput.Width = 15
	inputs = append(inputs, pidInput)
	labels = append(labels, "Process ID (PID)")

	return inputs, labels
}

// ========== FormData validation ==========

// ValidateForm validates form inputs for a specific command
func ValidateForm(command string, inputs []textinput.Model) error {
	switch command {
	case "config login":
		return validateLoginForm(inputs)
	case "config update":
		return validateConfigUpdateForm(inputs)
	case "exploit run", "exploit test":
		return validateExploitRunForm(inputs)
	case "exploit create", "exploit remove":
		return validateExploitNameForm(inputs)
	case "exploit stop":
		return validateExploitStopForm(inputs)
	}
	return nil
}

// validateLoginForm validates login form
func validateLoginForm(inputs []textinput.Model) error {
	if len(inputs) < 1 {
		return errors.New("invalid form structure")
	}

	password := strings.TrimSpace(inputs[0].Value())

	if password == "" {
		return errors.New("password is required")
	}

	return nil
}

// validateConfigUpdateForm validates config update form
func validateConfigUpdateForm(inputs []textinput.Model) error {
	if len(inputs) < 4 {
		return errors.New("invalid form structure")
	}

	host := strings.TrimSpace(inputs[0].Value())
	port := strings.TrimSpace(inputs[1].Value())
	username := strings.TrimSpace(inputs[2].Value())
	httpsStr := strings.TrimSpace(inputs[3].Value())

	if host == "" && port == "" && username == "" && httpsStr == "" {
		return errors.New("at least one field must be provided")
	}

	if httpsStr != "" && strings.ToLower(httpsStr) != "true" && strings.ToLower(httpsStr) != "false" {
		return errors.New("HTTPS field must be 'true' or 'false'")
	}

	return nil
}

// validateExploitRunForm validates exploit run form
func validateExploitRunForm(inputs []textinput.Model) error {
	if len(inputs) < 4 {
		return errors.New("invalid form structure")
	}

	exploitPath := strings.TrimSpace(inputs[0].Value())
	servicePort := strings.TrimSpace(inputs[1].Value())

	if exploitPath == "" {
		return errors.New("exploit path is required")
	}
	if servicePort == "" {
		return errors.New("service port is required")
	}

	return nil
}

// validateExploitNameForm validates exploit name form
func validateExploitNameForm(inputs []textinput.Model) error {
	if len(inputs) < 1 {
		return errors.New("invalid form structure")
	}

	name := strings.TrimSpace(inputs[0].Value())
	if name == "" {
		return errors.New("exploit name is required")
	}

	return nil
}

// validateExploitStopForm validates exploit stop form
// Skip validation since we're using a selection list
func validateExploitStopForm(inputs []textinput.Model) error {
	return nil
}

// GetFormData extracts data from form inputs
func GetFormData(inputs []textinput.Model, labels []string) FormData {
	data := FormData{Fields: make(map[string]string)}

	for i, input := range inputs {
		if i < len(labels) {
			data.Fields[labels[i]] = strings.TrimSpace(input.Value())
		}
	}

	return data
}

// ========== Form Navigation and Updates ==========

// NavigateForm handles navigation within forms
func NavigateForm(inputs []textinput.Model, currentFocus int, direction int) int {
	if len(inputs) == 0 {
		return 0
	}

	newFocus := currentFocus + direction
	if newFocus < 0 {
		newFocus = len(inputs) - 1
	} else if newFocus >= len(inputs) {
		newFocus = 0
	}

	for i := range inputs {
		if i == newFocus {
			inputs[i].Focus()
		} else {
			inputs[i].Blur()
		}
	}

	return newFocus
}

// UpdateFormInputs updates form inputs with a tea.Msg
func UpdateFormInputs(inputs []textinput.Model, focusIndex int, msg tea.Msg) ([]textinput.Model, tea.Cmd) {
	var cmd tea.Cmd

	if focusIndex >= 0 && focusIndex < len(inputs) {
		inputs[focusIndex], cmd = inputs[focusIndex].Update(msg)
	}

	return inputs, cmd
}
