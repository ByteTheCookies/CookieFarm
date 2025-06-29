package tui

import (
	"strconv"
	"strings"
	"time"

	"github.com/ByteTheCookies/CookieFarm/internal/client/config"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// New creates a new TUI model
func New(banner string) Model {
	mainMenu, configMenu, exploitMenu := InitializeMenus()

	spinnerInstance := spinner.New()
	spinnerInstance.Spinner = spinner.Points
	spinnerInstance.Style = SpinnerStyle

	exploitTable := table.New(
		table.WithColumns([]table.Column{
			{Title: "ID", Width: 6},
			{Title: "NAME", Width: 40},
			{Title: "PID", Width: 10},
		}),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	exploitTable.SetStyles(table.Styles{
		Header:   TableHeaderStyle,
		Selected: TableSelectedRowStyle,
		Cell:     TableRowStyle,
	})

	return Model{
		activeView:   "main",
		mainMenu:     mainMenu,
		configMenu:   configMenu,
		exploitMenu:  exploitMenu,
		banner:       banner,
		help:         help.New(),
		cmdRunner:    NewCommandRunner(),
		spinner:      spinnerInstance,
		exploitTable: exploitTable,
	}
}

// Init initializes the TUI
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// TableUpdateMsg is a message to update the table data
type TableUpdateMsg struct {
	Rows []table.Row
	Show bool
}

// Update handles TUI state updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if updateMsg, ok := msg.(TableUpdateMsg); ok {
		return m.handleTableUpdateMsg(updateMsg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case CommandOutput:
		return m.handleCommandOutput(msg)

	case ExploitOutput:
		return m.handleExploitOutput(msg)
	}

	if m.loading {
		cmds = append(cmds, m.updateSpinner(msg))
	}

	if m.showTable && (m.activeCommand == "exploit list" || m.activeCommand == "exploit stop") {
		cmds = append(cmds, m.updateExploitTable(msg))
	}

	cmds = append(cmds, m.updateActiveView(msg))
	return m, tea.Batch(cmds...)
}

func (m Model) handleTableUpdateMsg(updateMsg TableUpdateMsg) (tea.Model, tea.Cmd) {
	m.exploitTable.SetRows(updateMsg.Rows)
	m.showTable = updateMsg.Show
	return m, nil
}

func (m Model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width, m.height = msg.Width, msg.Height
	UpdateMenuSize(&m.mainMenu, &m.configMenu, &m.exploitMenu, msg.Width, msg.Height)
	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.IsInputMode() {
		return m.handleInputMode(msg)
	}

	if m.showTable && (m.activeCommand == "exploit list" || m.activeCommand == "exploit stop") {
		return m.handleTableKeyMsg(msg)
	}

	newM, cmdH, handled := m.handleKeyPress(msg)
	if handled {
		return newM, cmdH
	}

	return m.updateMenu(msg)
}

func (m Model) handleTableKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.activeCommand == "exploit stop" {
			handler := NewCommandHandler()
			m.SetLoading(true)
			return handler.ProcessFormSubmission(&m)
		}
		return m, nil
	case "esc", "q":
		return m.handleBackAction()
	}
	var cmd tea.Cmd
	m.exploitTable, cmd = m.exploitTable.Update(msg)
	return m, cmd
}

func (m Model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.GetActiveView() {
	case "main":
		m.mainMenu, cmd = m.mainMenu.Update(msg)
	case "config":
		m.configMenu, cmd = m.configMenu.Update(msg)
	case "exploit":
		m.exploitMenu, cmd = m.exploitMenu.Update(msg)
	}
	return m, cmd
}

func (m Model) handleCommandOutput(msg CommandOutput) (tea.Model, tea.Cmd) {
	m.SetCommandOutput(msg.Output)
	m.SetRunningCommand(true)
	m.SetLoading(false)
	if msg.Error != nil {
		m.SetError(msg.Error)
	}
	return m, nil
}

func (m Model) handleExploitOutput(msg ExploitOutput) (tea.Model, tea.Cmd) {
	if m.IsRunningCommand() && strings.HasPrefix(m.activeCommand, "exploit") {
		if msg.Content != "" {
			m.AppendCommandOutput(msg.Content)
		}
		if msg.Error != nil {
			m.SetError(msg.Error)
		}
		return m, tea.Batch(
			tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
				return nil
			}),
			m.cmdRunner.GetExploitOutputCmd(),
		)
	}
	return m, nil
}

func (m *Model) updateSpinner(msg tea.Msg) tea.Cmd {
	var spinnerCmd tea.Cmd
	m.spinner, spinnerCmd = m.spinner.Update(msg)
	return spinnerCmd
}

func (m *Model) updateExploitTable(msg tea.Msg) tea.Cmd {
	var tableCmd tea.Cmd
	m.exploitTable, tableCmd = m.exploitTable.Update(msg)
	return tableCmd
}

func (m Model) updateActiveView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch m.GetActiveView() {
	case "main":
		m.mainMenu, cmd = m.mainMenu.Update(msg)
	case "config":
		m.configMenu, cmd = m.configMenu.Update(msg)
	case "exploit":
		m.exploitMenu, cmd = m.exploitMenu.Update(msg)
	}
	return cmd
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	handler := NewCommandHandler()

	switch {
	case key.Matches(msg, Keys.Quit):
		m.quitting = true
		return m, tea.Quit, true

	case key.Matches(msg, Keys.Back):
		newM, cmd := m.handleBackAction()
		return newM, cmd, true

	case key.Matches(msg, Keys.Enter):
		newM, cmd := m.handleEnterAction(handler)
		return newM, cmd, true
	}

	return m, nil, false
}

// handleBackAction handles the back/escape key
func (m Model) handleBackAction() (tea.Model, tea.Cmd) {
	if m.showTable {
		m.showTable = false
		m.ClearError()

		if m.activeCommand == "exploit stop" {
			m.SetInputMode(false)
		}

		if !strings.HasPrefix(m.activeCommand, "exploit run") {
			m.SetRunningCommand(false)
		}

		return m, nil
	}

	if m.IsInputMode() {
		m.SetInputMode(false)
		m.ClearError()
		return m, nil
	}

	if m.IsRunningCommand() {
		m.SetRunningCommand(false)
		m.SetCommandOutput("")
		m.ClearError()
		return m, nil
	}

	if m.GetActiveView() != "main" {
		m.SetActiveView("main")
		return m, nil
	}

	return m, nil
}

// handleEnterAction handles the enter key
func (m Model) handleEnterAction(handler *CommandHandler) (tea.Model, tea.Cmd) {
	if m.IsInputMode() {
		m.SetLoading(true)
		newModel, cmd := handler.ProcessFormSubmission(&m)

		if strings.HasPrefix(newModel.activeCommand, "exploit run") {
			return newModel, tea.Batch(cmd, newModel.GetExploitStreamCmd(), m.spinner.Tick)
		}

		return newModel, tea.Batch(cmd, m.spinner.Tick)
	}

	selectedItem, ok := m.getSelectedMenuItem()
	if !ok {
		return m, nil
	}

	newModel, cmd := m.processMenuSelection(selectedItem, handler)

	if mTyped, ok := newModel.(Model); ok && mTyped.IsRunningCommand() {
		if strings.HasPrefix(mTyped.activeCommand, "exploit run") {
			return newModel, tea.Batch(cmd, mTyped.GetExploitStreamCmd())
		}
		if mTyped.activeCommand == "exploit list" {
			return newModel, tea.Batch(cmd, mTyped.SetupExploitTableCmd())
		}
	}

	return newModel, cmd
}

// getSelectedMenuItem gets the currently selected menu item
func (m Model) getSelectedMenuItem() (menuItem, bool) {
	switch m.GetActiveView() {
	case "main":
		return GetSelectedItem(m.mainMenu)
	case "config":
		return GetSelectedItem(m.configMenu)
	case "exploit":
		return GetSelectedItem(m.exploitMenu)
	default:
		return menuItem{}, false
	}
}

// processMenuSelection processes the selected menu item
func (m Model) processMenuSelection(selectedItem menuItem, handler *CommandHandler) (tea.Model, tea.Cmd) {
	command := selectedItem.command

	if IsNavigationCommand(command) {
		return handler.HandleNavigation(command, &m)
	}

	if IsDirectCommand(command) {
		m.activeCommand = command
		m.SetRunningCommand(true)
		m.SetLoading(true)

		if command == "exploit list" {
			return m, tea.Batch(
				handler.HandleCommand(command, nil),
				m.SetupExploitTableCmd(),
			)
		}

		return m, handler.HandleCommand(command, nil)
	}

	if RequiresInput(command) {
		handler.SetupFormForCommand(&m, command)
		return m, nil
	}

	return m, nil
}

// handleInputMode handles updates when in input mode
func (m Model) handleInputMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if m.IsProcessListVisible() && m.activeCommand == "exploit stop" {
			switch keyMsg.Type {
			case tea.KeyUp:
				m.SelectPreviousProcess()
				return m, nil
			case tea.KeyDown:
				m.SelectNextProcess()
				return m, nil
			case tea.KeyEnter:
				handler := NewCommandHandler()
				m.SetLoading(true)
				return handler.ProcessFormSubmission(&m)
			case tea.KeyEscape:
				m.SetInputMode(false)
				m.SetProcessListVisible(false)
				m.ClearError()
				m.activeCommand = ""
				m.SetActiveView("exploit")
				return m, nil
			}
			return m, nil
		}

		// Normal input handling
		switch keyMsg.Type {
		case tea.KeyTab:
			m.focusIndex = NavigateForm(m.inputs, m.focusIndex, 1)
			return m, nil
		case tea.KeyShiftTab:
			m.focusIndex = NavigateForm(m.inputs, m.focusIndex, -1)
			return m, nil
		case tea.KeyEnter:
			handler := NewCommandHandler()
			m.SetLoading(true)
			return handler.ProcessFormSubmission(&m)
		case tea.KeyEscape:
			m.SetInputMode(false)
			m.ClearError()
			return m, nil
		case tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight:
			m.inputs, cmd = UpdateFormInputs(m.inputs, m.focusIndex, msg)
			return m, cmd
		}
	}

	m.inputs, cmd = UpdateFormInputs(m.inputs, m.focusIndex, msg)
	return m, cmd
}

// View renders the TUI
func (m Model) View() string {
	renderer := NewViewRenderer()
	renderer.SetSize(m.width, m.height)
	return renderer.RenderView(&m)
}

// GetExploitStreamCmd returns a command that periodically checks for exploit output
func (m Model) GetExploitStreamCmd() tea.Cmd {
	if m.IsRunningCommand() && strings.HasPrefix(m.activeCommand, "exploit run") {
		return m.cmdRunner.GetExploitOutputCmd()
	}
	return nil
}

// SetupExploitTableCmd returns a command to update the exploit table with data
func (m Model) SetupExploitTableCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(100 * time.Millisecond)

		processes, err := m.cmdRunner.GetRunningExploits()
		if err != nil {
			return CommandOutput{
				Output: "Error fetching exploit data: " + err.Error(),
				Error:  err,
			}
		}

		if len(processes) == 0 {
			return CommandOutput{
				Output: "No running exploits found",
				Error:  nil,
			}
		}

		rows := []table.Row{}

		for _, p := range processes {
			rows = append(rows, table.Row{
				strconv.Itoa(p.ID),
				p.Name,
				strconv.Itoa(p.PID),
			})
		}

		modelCopy := m
		modelCopy.exploitTable.SetRows(rows)
		modelCopy.exploitTable.SetCursor(0)
		modelCopy.showTable = true

		return func() tea.Msg {
			return TableUpdateMsg{
				Rows: rows,
				Show: true,
			}
		}()
	}
}

// StartTUI launches the TUI application
func StartTUI(banner string) error {
	cm := config.GetConfigManager()
	err := cm.LoadLocalConfigFromFile()
	if err != nil {
		return err
	}

	logger.Setup("info", true)

	p := tea.NewProgram(
		New(banner),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err = p.Run()
	return err
}
