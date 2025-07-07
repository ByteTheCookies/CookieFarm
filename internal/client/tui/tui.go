package tui

import (
	"strconv"
	"time"

	"github.com/ByteTheCookies/CookieFarm/internal/client/config"
	"github.com/ByteTheCookies/CookieFarm/pkg/logger"
	"github.com/charmbracelet/bubbles/help"
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

// View renders the TUI
func (m Model) View() string {
	renderer := NewViewRenderer()
	renderer.SetSize(m.width, m.height)
	return renderer.RenderView(&m)
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
