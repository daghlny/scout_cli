package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/daghlny/scout_cli/pkg/engine"
)

// AITurnMsg triggers an AI player to take their turn.
type AITurnMsg struct {
	PlayerID int
}

// AIActionMsg carries the result of an AI's decision.
type AIActionMsg struct {
	PlayerID int
	Action   engine.Action
}

// StatusMsg sets a temporary status message.
type StatusMsg struct {
	Text    string
	IsError bool
}

// RoundEndMsg signals the round has ended.
type RoundEndMsg struct{}

// GameOverMsg signals the game has ended.
type GameOverMsg struct{}

// ClearStatusMsg clears the status message.
type ClearStatusMsg struct{}

// AIActionDoneMsg signals that an AI action's result has been displayed long enough.
type AIActionDoneMsg struct{}

// AIComputedMsg carries the result of an async AI computation (for LLM bots).
type AIComputedMsg struct {
	PlayerID int
	Action   engine.Action
	Error    string // non-empty if LLM failed (fallback was used)
}

// StartGameMsg starts a new game with the given config.
type StartGameMsg struct {
	NumPlayers int
	AIMode     string
}

// tickCmd returns a command that fires after a delay.
func tickCmd(d time.Duration, msg tea.Msg) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return msg
	})
}

// aiDelayCmd schedules an AI turn with a visual delay.
func aiDelayCmd(playerID int) tea.Cmd {
	return tickCmd(600*time.Millisecond, AITurnMsg{PlayerID: playerID})
}

func aiActionDoneCmd() tea.Cmd {
	return tickCmd(1200*time.Millisecond, AIActionDoneMsg{})
}
