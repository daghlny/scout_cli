package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/daghlny/scout_cli/pkg/engine"
)

// renderCard renders a single card as a styled box.
func renderCard(entry engine.HandEntry, style lipgloss.Style) string {
	active := activeValueStyle.Render(fmt.Sprintf("%2d", entry.ActiveValue()))
	sep := lipgloss.NewStyle().Foreground(colorPrimary).Render(" ♦ ")
	inactive := inactiveValueStyle.Render(fmt.Sprintf("%2d", entry.InactiveValue()))

	content := lipgloss.JoinVertical(lipgloss.Center, active, sep, inactive)
	return style.Width(5).Align(lipgloss.Center).Render(content)
}

type cardState int

const (
	cardNormal   cardState = iota
	cardCursor             // highlighted by cursor
	cardSelected           // selected for show
)

func cardStyle(state cardState) lipgloss.Style {
	switch state {
	case cardCursor:
		return cardCursorBorder
	case cardSelected:
		return cardSelectedBorder
	default:
		return cardNormalBorder
	}
}

// renderHand renders the player's hand with cursor and selected[]bool support.
func renderHand(hand *engine.Hand, cursor int, selected []bool) string {
	if hand.Len() == 0 {
		return lipgloss.NewStyle().Foreground(colorDim).Render("  (empty hand)")
	}

	entries := hand.Entries()
	cards := make([]string, len(entries))

	for i, e := range entries {
		state := cardNormal
		if len(selected) > i && selected[i] {
			state = cardSelected
		}
		if i == cursor {
			if state == cardSelected {
				// cursor on selected card: keep selected style
			} else {
				state = cardCursor
			}
		}
		cards[i] = renderCard(e, cardStyle(state))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cards...)
}

// renderHandWithInsert renders the hand with the scouted card preview at the insert position.
func renderHandWithInsert(hand *engine.Hand, insertPos int, scoutFromLeft bool, scoutFlip bool, table engine.TableState) string {
	// Build the scouted card entry for preview
	var scoutedEntry engine.HandEntry
	if len(table.Entries) > 0 {
		if scoutFromLeft {
			scoutedEntry = table.Entries[0]
		} else {
			scoutedEntry = table.Entries[len(table.Entries)-1]
		}
		if scoutFlip {
			scoutedEntry = scoutedEntry.FlippedEntry()
		}
	}

	preview := renderCard(scoutedEntry, cardInsertPreviewBorder)

	if hand.Len() == 0 {
		return preview
	}

	entries := hand.Entries()
	var parts []string

	for i, e := range entries {
		if i == insertPos {
			parts = append(parts, preview)
		}
		parts = append(parts, renderCard(e, cardNormalBorder))
	}
	// Insert at the end
	if insertPos >= hand.Len() {
		parts = append(parts, preview)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// renderTableCombo renders the current table combo.
func renderTableCombo(table engine.TableState, playerName string) string {
	if table.Combo == nil {
		return lipgloss.NewStyle().
			Foreground(colorDim).
			Italic(true).
			Render("  Table is empty — must Show")
	}

	cards := make([]string, len(table.Entries))
	for i, e := range table.Entries {
		cards[i] = renderCard(e, cardTableBorder)
	}

	comboRow := lipgloss.JoinHorizontal(lipgloss.Top, cards...)

	info := fmt.Sprintf("%s  %s (%d cards)  by %s",
		lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render("▶"),
		table.Combo.Type.String(),
		table.Combo.Size(),
		lipgloss.NewStyle().Foreground(colorAccent).Render(playerName),
	)

	return lipgloss.JoinVertical(lipgloss.Left, comboRow, info)
}

// renderTurnOrder renders a horizontal turn order indicator.
func renderTurnOrder(g *engine.GameState) string {
	var parts []string
	for i := 0; i < g.NumPlayers; i++ {
		// Order from StartPlayer
		pid := (g.StartPlayer + i) % g.NumPlayers
		p := g.Players[pid]
		name := p.Name
		isActive := pid == g.CurrentPlayer

		var label string
		if isActive {
			label = lipgloss.NewStyle().Foreground(colorGold).Bold(true).
				Render(fmt.Sprintf("▶ [%d] %s", i+1, name))
		} else {
			label = lipgloss.NewStyle().Foreground(colorDim).
				Render(fmt.Sprintf("  [%d] %s", i+1, name))
		}
		parts = append(parts, label)

		if i < g.NumPlayers-1 {
			parts = append(parts, lipgloss.NewStyle().Foreground(colorDim).Render(" → "))
		}
	}
	return strings.Join(parts, "")
}

// renderPlayersTable renders all players in a table format with S&S status.
// Uses lipgloss.Width() for correct column alignment with styled/emoji text.
func renderPlayersTable(g *engine.GameState) string {
	// Column widths
	const (
		colMark   = 2
		colNum    = 3
		colName   = 14
		colCards  = 7
		colToken  = 7
		colWon    = 7
		colSAS    = 5
	)

	dim := lipgloss.NewStyle().Foreground(colorDim)
	bold := lipgloss.NewStyle().Foreground(colorBright).Bold(true)

	markCol := lipgloss.NewStyle().Width(colMark)
	numCol := lipgloss.NewStyle().Width(colNum)
	nameCol := lipgloss.NewStyle().Width(colName)
	dataCol := lipgloss.NewStyle().Width(colCards).Align(lipgloss.Right)
	tokenCol := lipgloss.NewStyle().Width(colToken).Align(lipgloss.Right)
	wonCol := lipgloss.NewStyle().Width(colWon).Align(lipgloss.Right)
	sasCol := lipgloss.NewStyle().Width(colSAS).Align(lipgloss.Center)

	header := lipgloss.JoinHorizontal(lipgloss.Top,
		markCol.Render(""),
		numCol.Render(bold.Render("#")),
		nameCol.Render(bold.Render("Player")),
		dataCol.Render(bold.Render("Cards")),
		tokenCol.Render(bold.Render("Token")),
		wonCol.Render(bold.Render("Won")),
		sasCol.Render(bold.Render("S&S")),
	)

	totalWidth := colMark + colNum + colName + colCards + colToken + colWon + colSAS
	sep := dim.Render(strings.Repeat("─", totalWidth))

	var rows []string
	rows = append(rows, header)
	rows = append(rows, sep)

	for i := 0; i < g.NumPlayers; i++ {
		pid := (g.StartPlayer + i) % g.NumPlayers
		p := g.Players[pid]
		isActive := pid == g.CurrentPlayer

		nameStyle := lipgloss.NewStyle().Foreground(colorBright)
		numStyle := lipgloss.NewStyle().Foreground(colorBright)
		if isActive {
			nameStyle = nameStyle.Bold(true).Foreground(colorGold)
			numStyle = numStyle.Bold(true).Foreground(colorGold)
		}

		mark := ""
		if isActive {
			mark = lipgloss.NewStyle().Foreground(colorGold).Render("▶")
		}

		sas := lipgloss.NewStyle().Foreground(colorGreen).Render("Y")
		if p.UsedScoutShow {
			sas = lipgloss.NewStyle().Foreground(colorRed).Render("N")
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			markCol.Render(mark),
			numCol.Render(numStyle.Render(fmt.Sprintf("%d", i+1))),
			nameCol.Render(nameStyle.Render(p.Name)),
			dataCol.Render(fmt.Sprintf("%d", p.Hand.Len())),
			tokenCol.Render(fmt.Sprintf("%d", p.ScoutTokens)),
			wonCol.Render(fmt.Sprintf("%d", len(p.CollectedCards))),
			sasCol.Render(sas),
		)
		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// renderScoreboard renders a score summary table using fixed-width columns.
func renderScoreboard(players []engine.PlayerState, round int) string {
	const (
		colName  = 14
		colWon   = 7
		colToken = 7
		colHand  = 7
		colTotal = 7
	)

	bold := lipgloss.NewStyle().Bold(true)
	dim := lipgloss.NewStyle().Foreground(colorDim)
	nameCol := lipgloss.NewStyle().Width(colName)
	wonCol := lipgloss.NewStyle().Width(colWon).Align(lipgloss.Right)
	tokenCol := lipgloss.NewStyle().Width(colToken).Align(lipgloss.Right)
	handCol := lipgloss.NewStyle().Width(colHand).Align(lipgloss.Right)
	totalCol := lipgloss.NewStyle().Width(colTotal).Align(lipgloss.Right)

	header := lipgloss.JoinHorizontal(lipgloss.Top,
		nameCol.Render(bold.Foreground(colorBright).Render("Player")),
		wonCol.Render(bold.Foreground(colorGreen).Render("Won")),
		tokenCol.Render(bold.Foreground(colorAccent).Render("Token")),
		handCol.Render(bold.Foreground(colorRed).Render("Hand")),
		totalCol.Render(bold.Foreground(colorGold).Render("Total")),
	)

	totalWidth := colName + colWon + colToken + colHand + colTotal
	sep := dim.Render(strings.Repeat("─", totalWidth))

	var rows []string
	rows = append(rows, header)
	rows = append(rows, sep)

	for _, p := range players {
		score := p.TotalScore()
		var sStyle lipgloss.Style
		if score > 0 {
			sStyle = scorePositiveStyle
		} else if score < 0 {
			sStyle = scoreNegativeStyle
		} else {
			sStyle = scoreNeutralStyle
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			nameCol.Render(p.Name),
			wonCol.Render(fmt.Sprintf("%d", len(p.CollectedCards))),
			tokenCol.Render(fmt.Sprintf("%d", p.ScoutTokens)),
			handCol.Render(fmt.Sprintf("%d", p.Hand.Len())),
			totalCol.Render(sStyle.Render(fmt.Sprintf("%d", score))),
		)
		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
