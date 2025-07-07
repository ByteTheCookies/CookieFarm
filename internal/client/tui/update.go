package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// ======== HANDLES FUNCTIONS ========

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

// ======== UPDATE FUNCTIONS ========

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

// ======== OTHER UTILITY FUNCTIONS ========

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

// GetExploitStreamCmd returns a command that periodically checks for exploit output
func (m Model) GetExploitStreamCmd() tea.Cmd {
	if m.IsRunningCommand() && strings.HasPrefix(m.activeCommand, "exploit run") {
		return m.cmdRunner.GetExploitOutputCmd()
	}
	return nil
}
