package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MenuModel struct {
	cursor    int
	options   []menuOption
	width     int
	height    int
}

type menuOption struct {
	label      string
	numPlayers int
}

func NewMenuModel() MenuModel {
	return MenuModel{
		options: []menuOption{
			{"3 Players (You + 2 Bots)", 3},
			{"4 Players (You + 3 Bots)", 4},
			{"5 Players (You + 4 Bots)", 5},
		},
	}
}

func (m MenuModel) Update(msg tea.Msg) (MenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter", " ":
			return m, func() tea.Msg {
				return StartGameMsg{NumPlayers: m.options[m.cursor].numPlayers}
			}
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
	logo := `
   ╔═══╗ ╔═══╗ ╔═══╗ ╔╗ ╔╗ ╔════╗
   ║╔═╗║ ║╔═╗║ ║╔═╗║ ║║ ║║ ║╔╗╔╗║
   ║╚══╗ ║║ ╚╝ ║║ ║║ ║║ ║║ ╚╝║║╚╝
   ╚══╗║ ║║ ╔╗ ║║ ║║ ║║ ║║   ║║
   ║╚═╝║ ║╚═╝║ ║╚═╝║ ║╚═╝║   ║║
   ╚═══╝ ╚═══╝ ╚═══╝ ╚═══╝   ╚╝
`
	logoStyled := lipgloss.NewStyle().
		Foreground(colorGold).
		Bold(true).
		Render(logo)

	subtitle := lipgloss.NewStyle().
		Foreground(colorAccent).
		Italic(true).
		Render("       马 戏 星 探  ·  Card Game")

	var options []string
	for i, opt := range m.options {
		if i == m.cursor {
			prefix := lipgloss.NewStyle().Foreground(colorGold).Render("▸ ")
			options = append(options, menuSelectedStyle.Render(prefix+opt.label))
		} else {
			prefix := "  "
			options = append(options, menuItemStyle.Render(prefix+opt.label))
		}
	}
	optionList := lipgloss.JoinVertical(lipgloss.Left, options...)

	controls := fmt.Sprintf("\n  %s navigate  %s select  %s quit",
		keyStyle.Render("↑↓"),
		keyStyle.Render("Enter"),
		keyStyle.Render("q"),
	)

	content := lipgloss.JoinVertical(lipgloss.Center,
		logoStyled,
		subtitle,
		"",
		lipgloss.NewStyle().Foreground(colorBright).Bold(true).Render("  Select Players:"),
		"",
		optionList,
		descStyle.Render(controls),
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
