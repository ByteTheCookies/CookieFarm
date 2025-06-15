package tui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// CommandHandler manages command execution and handling
type CommandHandler struct {
	cmdRunner *CommandRunner
}

// NewCommandHandler creates a new command handler
func NewCommandHandler() *CommandHandler {
	return &CommandHandler{
		cmdRunner: NewCommandRunner(),
	}
}

// HandleCommand processes a command and returns appropriate tea.Cmd
func (h *CommandHandler) HandleCommand(command string, formData *FormData) tea.Cmd {
	if IsDirectCommand(command) {
		return h.executeDirectCommand(command)
	}

	if RequiresInput(command) && formData != nil {
		if command == "exploit stop" {
			return nil
		}
		return h.executeFormCommand(command, formData)
	}

	return nil
}

// executeDirectCommand executes commands that don't require input
func (h *CommandHandler) executeDirectCommand(command string) tea.Cmd {
	return tea.Sequence(
		tea.Println("Executing command..."),
		func() tea.Msg {
			parts := strings.Split(command, " ")
			if len(parts) != 2 {
				return CommandOutput{
					Output: "",
					Error:  fmt.Errorf("invalid command format: %s", command),
				}
			}

			var output string
			var err error

			switch parts[0] {
			case "config":
				output, err = h.cmdRunner.ExecuteConfigCommand(parts[1])
			case "exploit":
				output, err = h.cmdRunner.ExecuteExploitCommand(parts[1])
			default:
				err = fmt.Errorf("unknown command group: %s", parts[0])
			}

			return CommandOutput{
				Output: output,
				Error:  err,
			}
		},
	)
}

// executeFormCommand executes commands that require form input
func (h *CommandHandler) executeFormCommand(command string, formData *FormData) tea.Cmd {
	return tea.Sequence(
		tea.Println("Processing form data..."),
		func() tea.Msg {
			var output string
			var err error

			switch command {
			case "config login":
				output, err = h.handleLogin(formData)
			case "config update":
				output, err = h.handleConfigUpdate(formData)
			case "exploit run":
				output, err = h.handleExploitRun(formData)
			case "exploit test":
				output, err = h.handleExploitTest(formData)
			case "exploit create":
				output, err = h.handleExploitCreate(formData)
			case "exploit remove":
				output, err = h.handleExploitRemove(formData)
			case "exploit stop":
				output = "Exploit stop command executed"
			default:
				err = fmt.Errorf("unknown form command: %s", command)
			}

			return CommandOutput{
				Output: output,
				Error:  err,
			}
		},
	)
}

// handleLogin processes login command
func (h *CommandHandler) handleLogin(formData *FormData) (string, error) {
	password := formData.Fields["Password"]

	if password == "" {
		return "", errors.New("username and password are required")
	}

	return h.cmdRunner.ExecuteLogin(password)
}

// handleConfigUpdate processes config update command
func (h *CommandHandler) handleConfigUpdate(formData *FormData) (string, error) {
	host := formData.Fields["Host"]
	port := formData.Fields["Port"]
	username := formData.Fields["Username"]
	httpsStr := formData.Fields["HTTPS (true/false)"]

	useHTTPS := strings.ToLower(httpsStr) == "true"

	return h.cmdRunner.ExecuteConfigUpdate(host, port, username, useHTTPS)
}

// handleExploitRun processes exploit run command
func (h *CommandHandler) handleExploitRun(formData *FormData) (string, error) {
	exploitPath := formData.Fields["Exploit Path"]
	servicePort := formData.Fields["Service Port"]
	tickTime := formData.Fields["Tick Time (seconds)"]
	threadCount := formData.Fields["Thread Count"]

	if exploitPath == "" || servicePort == "" {
		return "", errors.New("exploit path and service port are required")
	}

	return h.cmdRunner.ExecuteExploitRun(exploitPath, servicePort, tickTime, threadCount)
}

// handleExploitRun processes exploit test command
func (h *CommandHandler) handleExploitTest(formData *FormData) (string, error) {
	exploitPath := formData.Fields["Exploit Path"]
	servicePort := formData.Fields["Service Port"]
	tickTime := formData.Fields["Tick Time (seconds)"]
	threadCount := formData.Fields["Thread Count"]

	if exploitPath == "" || servicePort == "" {
		return "", errors.New("exploit path and service port are required")
	}

	return h.cmdRunner.ExecuteExploitTest(exploitPath, servicePort, tickTime, threadCount)
}

// handleExploitCreate processes exploit create command
func (h *CommandHandler) handleExploitCreate(formData *FormData) (string, error) {
	name := formData.Fields["Exploit Name"]

	if name == "" {
		return "", errors.New("exploit name is required")
	}

	return h.cmdRunner.ExecuteExploitCreate(name)
}

// handleExploitRemove processes exploit remove command
func (h *CommandHandler) handleExploitRemove(formData *FormData) (string, error) {
	name := formData.Fields["Exploit Name"]

	if name == "" {
		return "", errors.New("exploit name is required")
	}

	return h.cmdRunner.ExecuteExploitRemove(name)
}

// handleExploitStop processes exploit stop command
func (h *CommandHandler) handleExploitStop(selectedProcess *ExploitProcess) (string, error) {
	if selectedProcess == nil {
		return "", errors.New("no exploit process selected")
	}

	pid := strconv.Itoa(selectedProcess.PID)

	result, err := h.cmdRunner.ExecuteExploitStop(pid)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Successfully stopped exploit: %s (PID: %d)\n\n%s",
		selectedProcess.Name, selectedProcess.PID, result), nil
}

// HandleNavigation processes navigation commands
func (*CommandHandler) HandleNavigation(command string, model *Model) (*Model, tea.Cmd) {
	switch command {
	case "quit":
		model.quitting = true
		return model, tea.Quit
	case "back":
		model.SetActiveView("main")
		model.SetInputMode(false)
		model.SetRunningCommand(false)
		model.ClearError()
		return model, nil
	case "config":
		model.SetActiveView("config")
		return model, nil
	case "exploit":
		model.SetActiveView("exploit")
		return model, nil
	}

	return model, nil
}

// ProcessFormSubmission handles form submission
func (h *CommandHandler) ProcessFormSubmission(model *Model) (*Model, tea.Cmd) {
	if err := ValidateForm(model.activeCommand, model.inputs); err != nil {
		model.SetError(err)
		return model, nil
	}

	formData := GetFormData(model.inputs, model.inputLabels)

	model.SetInputMode(false)
	model.SetRunningCommand(true)
	model.SetProcessListVisible(false)
	model.showTable = false
	model.ClearError()

	if model.activeCommand == "exploit stop" {
		var selectedProcess *ExploitProcess

		if model.showTable {
			selectedProcess = model.GetSelectedExploitFromTable()
		} else {
			selectedProcess = model.GetSelectedProcess()
		}

		return model, func() tea.Msg {
			output, err := h.handleExploitStop(selectedProcess)
			return CommandOutput{
				Output: output,
				Error:  err,
			}
		}
	}

	return model, h.HandleCommand(model.activeCommand, &formData)
}

// SetupFormForCommand prepares form inputs for a specific command
func (h *CommandHandler) SetupFormForCommand(model *Model, command string) {
	model.activeCommand = command
	model.inputs, model.inputLabels = CreateForm(command)
	model.focusIndex = 0
	model.SetInputMode(true)
	model.ClearError()

	if command == "exploit stop" {
		processes, err := h.cmdRunner.GetRunningExploits()
		if err != nil {
			model.SetError(err)
			return
		}

		if len(processes) == 0 {
			model.SetError(errors.New("no running exploits found"))
			return
		}

		var rows []table.Row
		for _, p := range processes {
			rows = append(rows, table.Row{
				strconv.Itoa(p.ID),
				p.Name,
				strconv.Itoa(p.PID),
			})
		}

		model.exploitTable.SetRows(rows)
		model.showTable = true

		model.SetExploitProcesses(processes)
		model.SetProcessListVisible(true)
	} else {
		model.showTable = false
		model.SetProcessListVisible(false)
	}
}
