package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daghlny/scout_cli/pkg/engine"
)

type GameOverModel struct {
	players  []engine.PlayerState
	winnerID int
	width    int
	height   int
}

func NewGameOverModel(g *engine.GameState) GameOverModel {
	return GameOverModel{
		players:  g.Players,
		winnerID: g.Winner(),
	}
}

func (m GameOverModel) Update(msg tea.Msg) (GameOverModel, tea.Cmd) {
	return m, nil
}

func (m GameOverModel) View() string {
	winner := m.players[m.winnerID]

	trophy := lipgloss.NewStyle().Foreground(colorGold).Bold(true).Render(`
    ╔═══════════════════════╗
    ║     🏆  GAME OVER     ║
    ╚═══════════════════════╝
`)

	winnerText := lipgloss.NewStyle().
		Foreground(colorGold).
		Bold(true).
		Render(fmt.Sprintf("    Winner: %s with %d points!", winner.Name, winner.TotalScore()))

	const (
		colName  = 14
		colRound = 6
		colTotal = 8
	)

	bold := lipgloss.NewStyle().Bold(true)
	dim := lipgloss.NewStyle().Foreground(colorDim)
	nameCol := lipgloss.NewStyle().Width(colName)
	roundCol := lipgloss.NewStyle().Width(colRound).Align(lipgloss.Right)
	totalColStyle := lipgloss.NewStyle().Width(colTotal).Align(lipgloss.Right)

	// Build header
	var headerParts []string
	headerParts = append(headerParts, nameCol.Render(bold.Foreground(colorBright).Render("Player")))
	numRounds := len(m.players[0].RoundScores)
	for i := 0; i < numRounds; i++ {
		headerParts = append(headerParts, roundCol.Render(bold.Foreground(colorPrimary).Render(fmt.Sprintf("R%d", i+1))))
	}
	headerParts = append(headerParts, totalColStyle.Render(bold.Foreground(colorGold).Render("Total")))
	header := lipgloss.JoinHorizontal(lipgloss.Top, headerParts...)

	totalWidth := colName + numRounds*colRound + colTotal
	sep := dim.Render(strings.Repeat("─", totalWidth))

	var rows []string
	rows = append(rows, header)
	rows = append(rows, sep)

	for _, p := range m.players {
		var parts []string
		parts = append(parts, nameCol.Render(p.Name))
		for _, s := range p.RoundScores {
			style := scoreNeutralStyle
			if s > 0 {
				style = scorePositiveStyle
			} else if s < 0 {
				style = scoreNegativeStyle
			}
			parts = append(parts, roundCol.Render(style.Render(fmt.Sprintf("%d", s))))
		}
		total := p.TotalScore()
		tStyle := scoreNeutralStyle
		if total > 0 {
			tStyle = scorePositiveStyle
		} else if total < 0 {
			tStyle = scoreNegativeStyle
		}
		suffix := ""
		if p.ID == m.winnerID {
			suffix = " 👑"
		}
		parts = append(parts, totalColStyle.Render(tStyle.Render(fmt.Sprintf("%d", total))+suffix))
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, parts...))
	}

	scoreTable := lipgloss.JoinVertical(lipgloss.Left, rows...)

	prompt := descStyle.Render(fmt.Sprintf("\n    Press %s for new game, %s to quit",
		keyStyle.Render("Enter"), keyStyle.Render("q")))

	content := lipgloss.JoinVertical(lipgloss.Left,
		trophy, winnerText, "", scoreTable, prompt,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
