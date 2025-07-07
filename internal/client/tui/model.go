package tui

import (
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
)

// Model is the main TUI model
type Model struct {
	err              error
	cmdRunner        *CommandRunner
	help             help.Model
	mainMenu         list.Model
	configMenu       list.Model
	exploitMenu      list.Model
	spinner          spinner.Model // Spinner for loading state
	exploitTable     table.Model   // Table for displaying exploit data
	inputLabels      []string
	inputs           []textinput.Model
	exploitProcesses []ExploitProcess // List of running exploit processes
	activeView       string
	banner           string
	activeCommand    string
	commandOutput    string
	lastUpdate       int64
	width, height    int
	focusIndex       int
	selectedProcess  int
	inputMode        bool
	runningCommand   bool
	quitting         bool
	streaming        bool // Whether streaming mode is active
	loading          bool // Whether a command is currently loading
	showProcessList  bool // Whether to show process selection list
	showTable        bool // Whether to show the table view
}

// TableUpdateMsg is a message to update the table data
type TableUpdateMsg struct {
	Rows []table.Row
	Show bool
}

// CommandOutput represents the result of a command execution
type CommandOutput struct {
	Output string
	Error  error
}

// menuItem represents an item in the menu
type menuItem struct {
	title       string
	description string
	command     string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.description }
func (i menuItem) FilterValue() string { return i.title }

// keyMap defines keybindings
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Quit      key.Binding
	Back      key.Binding
	Tab       key.Binding
	Submit    key.Binding
	NextInput key.Binding
	PrevInput key.Binding
}

// Keys defines the keybindings
var Keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next input"),
	),
	Submit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	),
	NextInput: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next input"),
	),
	PrevInput: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous input"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Quit, k.Back},
	}
}

// ========== Model Methods ==========

// SetError sets an error message for the model
func (m *Model) SetError(err error) {
	m.err = err
}

// ClearError clears any error message
func (m *Model) ClearError() {
	m.err = nil
}

// IsInputMode returns true if the model is in input mode
func (m *Model) IsInputMode() bool {
	return m.inputMode
}

// SetInputMode sets the input mode state
func (m *Model) SetInputMode(state bool) {
	m.inputMode = state
}

// IsRunningCommand returns true if a command is currently running
func (m *Model) IsRunningCommand() bool {
	return m.runningCommand
}

// SetRunningCommand sets the running command state
func (m *Model) SetRunningCommand(state bool) {
	m.runningCommand = state

	// When stopping command execution, also stop streaming and loading
	if !state {
		m.streaming = false
		m.loading = false
	}
}

// SetCommandOutput sets the command output
func (m *Model) SetCommandOutput(output string) {
	m.commandOutput = output
}

// AppendCommandOutput appends to the existing command output
func (m *Model) AppendCommandOutput(output string) {
	if output == "" {
		return
	}

	if m.commandOutput == "" {
		m.commandOutput = output
	} else {
		m.commandOutput += "\n" + output
	}

	// Update streaming timestamp
	if m.streaming {
		m.lastUpdate = time.Now().UnixNano()
	}
}

// GetActiveView returns the current active view
func (m *Model) GetActiveView() string {
	return m.activeView
}

// SetActiveView sets the active view
func (m *Model) SetActiveView(view string) {
	m.activeView = view
}

// SetLoading sets the loading state
func (m *Model) SetLoading(state bool) {
	m.loading = state
}

// IsLoading returns true if the model is in loading state
func (m *Model) IsLoading() bool {
	return m.loading
}

// GetSelectedTableRow returns the currently selected table row
func (m *Model) GetSelectedTableRow() table.Row {
	return m.exploitTable.SelectedRow()
}

// GetSelectedExploitFromTable returns the exploit process from the selected table row
func (m *Model) GetSelectedExploitFromTable() *ExploitProcess {
	row := m.exploitTable.SelectedRow()
	if len(row) < 3 {
		return nil
	}

	id, err := strconv.Atoi(row[0])
	if err != nil {
		return nil
	}

	pid, err := strconv.Atoi(row[2])
	if err != nil {
		return nil
	}

	return &ExploitProcess{
		ID:   id,
		Name: row[1],
		PID:  pid,
	}
}

// SetExploitProcesses sets the list of running exploit processes
func (m *Model) SetExploitProcesses(processes []ExploitProcess) {
	m.exploitProcesses = processes
	m.selectedProcess = 0 // Reset selection
}

// GetSelectedProcess returns the currently selected exploit process
func (m *Model) GetSelectedProcess() *ExploitProcess {
	if len(m.exploitProcesses) == 0 || m.selectedProcess < 0 || m.selectedProcess >= len(m.exploitProcesses) {
		return nil
	}
	return &m.exploitProcesses[m.selectedProcess]
}

// SetProcessListVisible shows or hides the process selection list
func (m *Model) SetProcessListVisible(visible bool) {
	m.showProcessList = visible
}

// IsProcessListVisible returns whether the process list is visible
func (m *Model) IsProcessListVisible() bool {
	return m.showProcessList
}

// SelectNextProcess selects the next process in the list
func (m *Model) SelectNextProcess() {
	if len(m.exploitProcesses) == 0 {
		return
	}
	m.selectedProcess = (m.selectedProcess + 1) % len(m.exploitProcesses)
}

// SelectPreviousProcess selects the previous process in the list
func (m *Model) SelectPreviousProcess() {
	if len(m.exploitProcesses) == 0 {
		return
	}
	m.selectedProcess = (m.selectedProcess - 1 + len(m.exploitProcesses)) % len(m.exploitProcesses)
}

// ========== Exploit Process Management ==========

// ExploitOutput represents output from a running exploit
type ExploitOutput struct {
	Content string
	Error   error
	PID     int
}

// CommandRunner handles the execution of commands for the TUI
type CommandRunner struct {
	exploitOutputChan chan ExploitOutput
	currentExploitPID int
}

// ExploitProcess represents a running exploit process
type ExploitProcess struct {
	ID   int
	Name string
	PID  int
}

// ExploitTableData holds exploit process data
var ExploitTableData []ExploitProcess

// ========== Command Execution ==========

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

// ========== Form models ==========

// FormData represents form input data
type FormData struct {
	Fields map[string]string
}
