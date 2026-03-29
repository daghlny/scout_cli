package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daghlny/scout_cli/pkg/engine"
)

type RoundEndModel struct {
	players      []engine.PlayerState
	round        int
	totalRounds  int
	roundEnderID int
	width        int
	height       int
}

func NewRoundEndModel(g *engine.GameState) RoundEndModel {
	return RoundEndModel{
		players:      g.Players,
		round:        g.Round,
		totalRounds:  g.TotalRounds,
		roundEnderID: g.RoundEnderID,
	}
}

func (m RoundEndModel) Update(msg tea.Msg) (RoundEndModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "enter" || msg.String() == " " {
			return m, nil // handled by parent
		}
	}
	return m, nil
}

func (m RoundEndModel) View() string {
	title := lipgloss.NewStyle().
		Foreground(colorGold).
		Bold(true).
		Render(fmt.Sprintf("══════ Round %d Complete ══════", m.round+1))

	enderName := m.players[m.roundEnderID].Name
	enderInfo := lipgloss.NewStyle().
		Foreground(colorAccent).
		Render(fmt.Sprintf("  Round ended by: %s", enderName))

	scoreboard := renderScoreboard(m.players, m.round)

	prompt := descStyle.Render(fmt.Sprintf("\n  Press %s to continue", keyStyle.Render("Enter")))

	content := lipgloss.JoinVertical(lipgloss.Left,
		"", title, "", enderInfo, "", scoreboard, prompt,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
