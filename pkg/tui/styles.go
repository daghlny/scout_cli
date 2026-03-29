package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorPrimary   = lipgloss.Color("#7D56F4")
	colorSecondary = lipgloss.Color("#FF6B6B")
	colorAccent    = lipgloss.Color("#4ECDC4")
	colorGold      = lipgloss.Color("#FFD700")
	colorDim       = lipgloss.Color("#666666")
	colorBright    = lipgloss.Color("#FFFFFF")
	colorBg        = lipgloss.Color("#1A1A2E")
	colorCardBg    = lipgloss.Color("#16213E")
	colorGreen     = lipgloss.Color("#00E676")
	colorRed       = lipgloss.Color("#FF5252")
	colorBlue      = lipgloss.Color("#448AFF")

	// Title
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorGold).
			MarginBottom(1)

	// Card styles
	cardNormalBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorPrimary).
				Padding(0, 1)

	cardSelectedBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorGold).
				Padding(0, 1)

	cardCursorBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorAccent).
				Padding(0, 1)

	cardTableBorder = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(colorGreen).
			Padding(0, 1)

	// Values
	activeValueStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorBright)

	inactiveValueStyle = lipgloss.NewStyle().
				Foreground(colorDim)

	// Info panels
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDim).
			Padding(0, 1)

	// Action bar
	actionBarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorDim).
			Padding(0, 1).
			MarginTop(1)

	// Status message
	statusStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Italic(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	// Menu
	menuItemStyle = lipgloss.NewStyle().
			Foreground(colorBright).
			PaddingLeft(2)

	menuSelectedStyle = lipgloss.NewStyle().
				Foreground(colorGold).
				Bold(true).
				PaddingLeft(2)

	// Opponent card back (compact)
	cardBackStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	// Score styles
	scorePositiveStyle = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	scoreNegativeStyle = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	scoreNeutralStyle  = lipgloss.NewStyle().Foreground(colorBright)

	// Key hint
	keyStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	descStyle = lipgloss.NewStyle().
			Foreground(colorDim)
)
