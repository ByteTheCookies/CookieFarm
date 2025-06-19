package tui

import (
	"github.com/charmbracelet/bubbles/list"
)

// InitializeMenus initializes all the menus for the TUI
func InitializeMenus() (list.Model, list.Model, list.Model) {
	mainMenu := createMainMenu()
	configMenu := createConfigMenu()
	exploitMenu := createExploitMenu()

	return mainMenu, configMenu, exploitMenu
}

// createMainMenu creates the main menu
func createMainMenu() list.Model {
	mainMenuItems := []list.Item{
		menuItem{title: "Configuration", description: "Manage client configuration", command: "config"},
		menuItem{title: "Exploits", description: "Manage and run exploits", command: "exploit"},
		menuItem{title: "Quit", description: "Exit the application", command: "quit"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles = ListStyle

	mainMenuList := list.New(mainMenuItems, delegate, 0, 0)
	mainMenuList.Title = "CookieFarm Client"
	mainMenuList.SetShowStatusBar(false)
	mainMenuList.SetFilteringEnabled(false)
	mainMenuList.Styles.Title = MenuTitleStyle

	return mainMenuList
}

// createConfigMenu creates the configuration menu
func createConfigMenu() list.Model {
	configMenuItems := []list.Item{
		menuItem{title: "Login", description: "Login to the server", command: "config login"},
		menuItem{title: "Show Config", description: "Show current configuration", command: "config show"},
		menuItem{title: "Update Config", description: "Update configuration settings", command: "config update"},
		menuItem{title: "Reset Config", description: "Reset configuration to defaults", command: "config reset"},
		menuItem{title: "Logout", description: "Logout from the server", command: "config logout"},
		menuItem{title: "Back", description: "Return to main menu", command: "back"},
	}
	delegate := list.NewDefaultDelegate()
	delegate.Styles = ListStyle
	configMenuList := list.New(configMenuItems, delegate, 0, 0)
	configMenuList.Title = "Configuration Menu"
	configMenuList.SetShowStatusBar(false)
	configMenuList.SetFilteringEnabled(false)
	configMenuList.Styles.Title = MenuTitleStyle

	return configMenuList
}

// createExploitMenu creates the exploit menu
func createExploitMenu() list.Model {
	exploitMenuItems := []list.Item{
		menuItem{title: "Run Exploit", description: "Run an exploit against other teams", command: "exploit run"},
		menuItem{title: "Test Exploit", description: "Test an exploit on the nop team", command: "exploit test"},
		menuItem{title: "Create Exploit", description: "Create a new exploit template", command: "exploit create"},
		menuItem{title: "List Exploits", description: "List all running exploits", command: "exploit list"},
		menuItem{title: "Stop Exploit", description: "Stop a running exploit", command: "exploit stop"},
		menuItem{title: "Remove Exploit", description: "Remove an exploit template", command: "exploit remove"},
		menuItem{title: "Back", description: "Return to main menu", command: "back"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles = ListStyle
	exploitMenuList := list.New(exploitMenuItems, delegate, 0, 0)
	exploitMenuList.Title = "Exploit Menu"
	exploitMenuList.SetShowStatusBar(false)
	exploitMenuList.SetFilteringEnabled(false)
	exploitMenuList.Styles.Title = MenuTitleStyle

	return exploitMenuList
}

// GetSelectedItem returns the selected item from the given menu
func GetSelectedItem(menu list.Model) (menuItem, bool) {
	if item := menu.SelectedItem(); item != nil {
		if menuItem, ok := item.(menuItem); ok {
			return menuItem, true
		}
	}
	return menuItem{}, false
}

// UpdateMenuSize updates the size of all menus
func UpdateMenuSize(mainMenu, configMenu, exploitMenu *list.Model, width, height int) {
	headerHeight := 12 // Banner + title
	footerHeight := 2  // Help section

	menuHeight := height - headerHeight - footerHeight

	mainMenu.SetWidth(width)
	mainMenu.SetHeight(menuHeight)

	configMenu.SetWidth(width)
	configMenu.SetHeight(menuHeight)

	exploitMenu.SetWidth(width)
	exploitMenu.SetHeight(menuHeight)
}

// IsNavigationCommand checks if the command is a navigation command
func IsNavigationCommand(command string) bool {
	switch command {
	case "quit", "back", "config", "exploit":
		return true
	default:
		return false
	}
}

// IsDirectCommand checks if the command can be executed directly without input
func IsDirectCommand(command string) bool {
	switch command {
	case "config show", "config reset", "config logout", "exploit list":
		return true
	default:
		return false
	}
}

// RequiresInput checks if the command requires user input
func RequiresInput(command string) bool {
	switch command {
	case "config login", "config update", "exploit run", "exploit create", "exploit remove", "exploit stop", "exploit test":
		return true
	default:
		return false
	}
}
