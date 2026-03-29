package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Screen int

const (
	ScreenMenu     Screen = iota
	ScreenGame
	ScreenRoundEnd
	ScreenGameOver
)

type AppModel struct {
	screen        Screen
	menuModel     MenuModel
	gameModel     GameModel
	roundEndModel RoundEndModel
	gameOverModel GameOverModel
	width         int
	height        int
}

func NewApp() AppModel {
	return AppModel{
		screen:    ScreenMenu,
		menuModel: NewMenuModel(),
	}
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.menuModel.width = msg.Width
		m.menuModel.height = msg.Height
		m.gameModel.width = msg.Width
		m.gameModel.height = msg.Height
		m.roundEndModel.width = msg.Width
		m.roundEndModel.height = msg.Height
		m.gameOverModel.width = msg.Width
		m.gameOverModel.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "q" && (m.screen == ScreenMenu || m.screen == ScreenGameOver) {
			return m, tea.Quit
		}
	}

	switch m.screen {
	case ScreenMenu:
		return m.updateMenu(msg)
	case ScreenGame:
		return m.updateGame(msg)
	case ScreenRoundEnd:
		return m.updateRoundEnd(msg)
	case ScreenGameOver:
		return m.updateGameOver(msg)
	}

	return m, nil
}

func (m AppModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StartGameMsg:
		gm, cmd := NewGameModel(msg.NumPlayers)
		gm.width = m.width
		gm.height = m.height
		m.gameModel = gm
		m.screen = ScreenGame
		return m, cmd
	default:
		var cmd tea.Cmd
		m.menuModel, cmd = m.menuModel.Update(msg)
		return m, cmd
	}
}

func (m AppModel) updateGame(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case RoundEndMsg:
		m.roundEndModel = NewRoundEndModel(m.gameModel.engine)
		m.roundEndModel.width = m.width
		m.roundEndModel.height = m.height
		m.screen = ScreenRoundEnd
		return m, nil
	case GameOverMsg:
		m.gameOverModel = NewGameOverModel(m.gameModel.engine)
		m.gameOverModel.width = m.width
		m.gameOverModel.height = m.height
		m.screen = ScreenGameOver
		return m, nil
	default:
		var cmd tea.Cmd
		m.gameModel, cmd = m.gameModel.Update(msg)
		return m, cmd
	}
}

func (m AppModel) updateRoundEnd(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "enter" || keyMsg.String() == " " {
			if err := m.gameModel.engine.NextRound(); err == nil {
				if m.gameModel.engine.Phase == 3 { // PhaseGameEnd
					m.gameOverModel = NewGameOverModel(m.gameModel.engine)
					m.gameOverModel.width = m.width
					m.gameOverModel.height = m.height
					m.screen = ScreenGameOver
					return m, nil
				}
				m.gameModel.uiPhase = UIPhaseFlipChoice
				m.gameModel.cursor = 0
				m.gameModel.selected = nil
				m.gameModel.statusMsg = ""
				m.gameModel.flipCount = 0
				m.gameModel.isScoutAndShow = false
				m.screen = ScreenGame
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.roundEndModel, cmd = m.roundEndModel.Update(msg)
	return m, cmd
}

func (m AppModel) updateGameOver(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "enter" {
			m.menuModel = NewMenuModel()
			m.menuModel.width = m.width
			m.menuModel.height = m.height
			m.screen = ScreenMenu
			return m, nil
		}
	}
	return m, nil
}

func (m AppModel) View() string {
	switch m.screen {
	case ScreenMenu:
		return m.menuModel.View()
	case ScreenGame:
		return m.gameModel.View()
	case ScreenRoundEnd:
		return m.roundEndModel.View()
	case ScreenGameOver:
		return m.gameOverModel.View()
	}
	return ""
}
