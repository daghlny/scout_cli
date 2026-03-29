package tui

import (
	"fmt"
	"math/rand"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daghlny/scout_cli/pkg/ai"
	"github.com/daghlny/scout_cli/pkg/engine"
)

// UIPhase tracks the multi-step human interaction flow.
type UIPhase int

const (
	UIPhaseFlipChoice     UIPhase = iota // choose hand orientation
	UIPhaseChooseAction                  // pick Show/Scout/Scout&Show
	UIPhaseShowSelect                    // select adjacent cards to show
	UIPhaseScoutEnd                      // choose left or right end of table
	UIPhaseScoutFlip                     // choose whether to flip scouted card
	UIPhaseScoutInsert                   // choose insert position
	UIPhaseWaitingForAI                  // waiting for AI delay
)

type GameModel struct {
	engine  *engine.GameState
	bots    []ai.Strategy
	humanID int

	// UI state
	uiPhase       UIPhase
	cursor        int
	selected      []bool // per-card selection for Show
	actionCursor  int    // for action menu
	scoutFromLeft bool
	scoutFlip     bool
	flipCount     int  // tracks flip previews during setup
	isScoutAndShow bool // true when doing Scout & Show combined action
	// saved scout params for Scout & Show
	savedScoutFromLeft bool
	savedScoutFlip     bool
	savedScoutInsertAt int
	statusMsg     string
	statusIsError bool
	width         int
	height        int
}

func NewGameModel(numPlayers int) (GameModel, tea.Cmd) {
	rng := rand.New(rand.NewSource(rand.Int63()))

	names := make([]string, numPlayers)
	names[0] = "You"
	botNames := []string{"Alice 🤖", "Bob 🤖", "Carol 🤖", "Dave 🤖"}
	for i := 1; i < numPlayers; i++ {
		names[i] = botNames[i-1]
	}

	g, _ := engine.NewGame(numPlayers, names, rng)

	bots := make([]ai.Strategy, numPlayers)
	for i := 1; i < numPlayers; i++ {
		bots[i] = ai.NewSmartBot(rand.New(rand.NewSource(rng.Int63())))
	}

	m := GameModel{
		engine:  g,
		bots:    bots,
		humanID: 0,
		uiPhase: UIPhaseFlipChoice,
		cursor:  0,
	}

	return m, nil
}

func (m GameModel) Update(msg tea.Msg) (GameModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case AITurnMsg:
		return m.handleAITurn(msg)
	case ClearStatusMsg:
		m.statusMsg = ""
		m.statusIsError = false
		return m, nil
	}
	return m, nil
}

func (m GameModel) handleKey(msg tea.KeyMsg) (GameModel, tea.Cmd) {
	key := msg.String()

	// Don't accept input while waiting for AI
	if m.uiPhase == UIPhaseWaitingForAI {
		return m, nil
	}

	switch m.uiPhase {
	case UIPhaseFlipChoice:
		return m.handleFlipChoice(key)
	case UIPhaseChooseAction:
		return m.handleChooseAction(key)
	case UIPhaseShowSelect:
		return m.handleShowSelect(key)
	case UIPhaseScoutEnd:
		return m.handleScoutEnd(key)
	case UIPhaseScoutFlip:
		return m.handleScoutFlip(key)
	case UIPhaseScoutInsert:
		return m.handleScoutInsert(key)
	}
	return m, nil
}

func (m GameModel) handleFlipChoice(key string) (GameModel, tea.Cmd) {
	switch key {
	case "f":
		// Preview flip: toggle the hand visually
		m.engine.Players[m.humanID].Hand.FlipAll()
		m.flipCount++
		m.statusMsg = "Preview: hand flipped. Press [F] again or [Enter] to confirm."
		return m, nil
	case "enter":
		// Confirm: the hand is already in the desired state from preview flips.
		// Tell the engine the final orientation.
		flip := m.flipCount%2 == 1
		// Undo the visual flips so SetHandOrientation can apply cleanly
		if flip {
			m.engine.Players[m.humanID].Hand.FlipAll()
		}
		_ = m.engine.SetHandOrientation(m.humanID, flip)
		m.flipCount = 0
		return m.finishSetup()
	}
	return m, nil
}

func (m GameModel) finishSetup() (GameModel, tea.Cmd) {
	// AI bots choose orientation
	for i, bot := range m.bots {
		if bot == nil {
			continue
		}
		flip := bot.ChooseHandOrientation(m.engine.Players[i].Hand, m.engine)
		_ = m.engine.SetHandOrientation(i, flip)
	}
	_ = m.engine.FinishSetup()
	m.uiPhase = UIPhaseChooseAction

	// If AI goes first
	if m.engine.CurrentPlayer != m.humanID {
		m.uiPhase = UIPhaseWaitingForAI
		return m, aiDelayCmd(m.engine.CurrentPlayer)
	}

	// If table is empty, human must show
	if m.engine.Table.Combo == nil {
		m = m.enterShowSelect()
		m.statusMsg = "Table is empty — you must Show."
	}
	return m, nil
}

func (m GameModel) enterShowSelect() GameModel {
	hand := m.engine.Players[m.humanID].Hand
	m.uiPhase = UIPhaseShowSelect
	m.cursor = 0
	m.selected = make([]bool, hand.Len())
	m.statusMsg = "←→ move, Space select/deselect, Enter play, Esc back"
	return m
}

func (m GameModel) handleChooseAction(key string) (GameModel, tea.Cmd) {
	tableEmpty := m.engine.Table.Combo == nil
	if tableEmpty {
		m = m.enterShowSelect()
		return m, nil
	}

	switch key {
	case "1", "s":
		m.isScoutAndShow = false
		m = m.enterShowSelect()
		return m, nil
	case "2", "c":
		m.isScoutAndShow = false
		m.uiPhase = UIPhaseScoutEnd
		m.statusMsg = "Scout: press ← for left card, → for right card."
		return m, nil
	case "3", "b":
		player := &m.engine.Players[m.humanID]
		if player.UsedScoutShow {
			m.statusMsg = "Scout & Show already used this round!"
			m.statusIsError = true
			return m, tickCmd(2000*1000000, ClearStatusMsg{})
		}
		m.isScoutAndShow = true
		m.uiPhase = UIPhaseScoutEnd
		m.statusMsg = "Scout & Show: first, choose which end to scout."
		return m, nil
	}
	return m, nil
}

func (m GameModel) handleShowSelect(key string) (GameModel, tea.Cmd) {
	hand := m.engine.Players[m.humanID].Hand

	switch key {
	case "left", "h":
		if m.cursor > 0 {
			m.cursor--
		}
	case "right", "l":
		if m.cursor < hand.Len()-1 {
			m.cursor++
		}
	case " ":
		// Toggle selection at cursor
		if m.cursor >= 0 && m.cursor < len(m.selected) {
			m.selected[m.cursor] = !m.selected[m.cursor]
		}
	case "enter":
		return m.executeShow()
	case "escape":
		if m.isScoutAndShow {
			// Undo the simulated insert and cancel
			m.undoSimulatedScout()
			m.isScoutAndShow = false
			m.uiPhase = UIPhaseChooseAction
			m.selected = nil
			m.statusMsg = ""
			return m, nil
		}
		if m.engine.Table.Combo != nil {
			m.uiPhase = UIPhaseChooseAction
			m.selected = nil
			m.statusMsg = ""
		}
		return m, nil
	}
	return m, nil
}

// getScoutedEntry returns the HandEntry that would be scouted based on current scout params.
func (m GameModel) getScoutedEntry() engine.HandEntry {
	var entry engine.HandEntry
	if m.scoutFromLeft {
		entry = m.engine.Table.Entries[0]
	} else {
		entry = m.engine.Table.Entries[len(m.engine.Table.Entries)-1]
	}
	if m.scoutFlip {
		entry = entry.FlippedEntry()
	}
	return entry
}

// undoSimulatedScout removes the card that was visually inserted during Scout & Show.
func (m *GameModel) undoSimulatedScout() {
	m.engine.Players[m.humanID].Hand.RemoveRange(m.savedScoutInsertAt, 1)
}

// selectedIndices returns the sorted indices of selected cards.
func (m GameModel) selectedIndices() []int {
	var indices []int
	for i, sel := range m.selected {
		if sel {
			indices = append(indices, i)
		}
	}
	return indices
}

func (m GameModel) executeShow() (GameModel, tea.Cmd) {
	indices := m.selectedIndices()
	if len(indices) == 0 {
		m.statusMsg = "No cards selected!"
		m.statusIsError = true
		return m, tickCmd(2000*1000000, ClearStatusMsg{})
	}

	// Check continuity
	for i := 1; i < len(indices); i++ {
		if indices[i] != indices[i-1]+1 {
			m.statusMsg = "Selected cards must be adjacent!"
			m.statusIsError = true
			return m, tickCmd(2000*1000000, ClearStatusMsg{})
		}
	}

	showStart := indices[0]
	showCount := len(indices)

	if m.isScoutAndShow {
		// Undo the simulated insert so the engine can apply ScoutAndShow atomically
		m.undoSimulatedScout()

		action := engine.Action{
			Type:          engine.ActionScoutAndShow,
			ScoutFromLeft: m.savedScoutFromLeft,
			ScoutFlip:     m.savedScoutFlip,
			ScoutInsertAt: m.savedScoutInsertAt,
			ShowStart:     showStart,
			ShowCount:     showCount,
		}
		if err := m.engine.ApplyAction(action); err != nil {
			// Re-insert the card so user can retry selection
			entry := m.getScoutedEntry()
			_ = m.engine.Players[m.humanID].Hand.Insert(entry, m.savedScoutInsertAt)
			m.statusMsg = "Invalid: " + err.Error()
			m.statusIsError = true
			return m, tickCmd(2000*1000000, ClearStatusMsg{})
		}
		m.isScoutAndShow = false
	} else {
		action := engine.Action{
			Type:      engine.ActionShow,
			ShowStart: showStart,
			ShowCount: showCount,
		}
		if err := m.engine.ApplyAction(action); err != nil {
			m.statusMsg = "Invalid: " + err.Error()
			m.statusIsError = true
			return m, tickCmd(2000*1000000, ClearStatusMsg{})
		}
	}

	m.selected = nil
	m.statusMsg = ""
	return m.afterAction()
}

func (m GameModel) handleScoutEnd(key string) (GameModel, tea.Cmd) {
	switch key {
	case "left", "h":
		m.scoutFromLeft = true
		m.uiPhase = UIPhaseScoutFlip
		m.statusMsg = "Flip the card? (f)lip / (k)eep"
		return m, nil
	case "right", "l":
		m.scoutFromLeft = false
		m.uiPhase = UIPhaseScoutFlip
		m.statusMsg = "Flip the card? (f)lip / (k)eep"
		return m, nil
	case "escape", "q":
		m.uiPhase = UIPhaseChooseAction
		m.statusMsg = ""
		return m, nil
	}
	return m, nil
}

func (m GameModel) handleScoutFlip(key string) (GameModel, tea.Cmd) {
	switch key {
	case "f":
		m.scoutFlip = true
	case "k", "enter":
		m.scoutFlip = false
	case "escape", "q":
		m.uiPhase = UIPhaseScoutEnd
		m.statusMsg = "Scout: press ← for left card, → for right card."
		return m, nil
	default:
		return m, nil
	}

	m.uiPhase = UIPhaseScoutInsert
	m.cursor = 0
	m.statusMsg = "Choose insert position with ←→, then Enter."
	return m, nil
}

func (m GameModel) handleScoutInsert(key string) (GameModel, tea.Cmd) {
	hand := m.engine.Players[m.humanID].Hand

	switch key {
	case "left", "h":
		if m.cursor > 0 {
			m.cursor--
		}
	case "right", "l":
		if m.cursor < hand.Len() {
			m.cursor++
		}
	case "enter":
		return m.executeScout()
	case "escape", "q":
		m.uiPhase = UIPhaseScoutFlip
		m.statusMsg = "Flip the card? (f)lip / (k)eep"
		return m, nil
	}
	return m, nil
}

func (m GameModel) executeScout() (GameModel, tea.Cmd) {
	if m.isScoutAndShow {
		// Save scout params for the final ScoutAndShow action.
		m.savedScoutFromLeft = m.scoutFromLeft
		m.savedScoutFlip = m.scoutFlip
		m.savedScoutInsertAt = m.cursor

		// Simulate inserting the scouted card into hand so the user can see
		// and select it during the Show phase. The engine's ApplyAction for
		// ScoutAndShow will handle the real state atomically.
		entry := m.getScoutedEntry()
		_ = m.engine.Players[m.humanID].Hand.Insert(entry, m.cursor)

		m.statusMsg = "Card scouted! Now select cards to Show."
		m = m.enterShowSelect()
		return m, nil
	}

	action := engine.Action{
		Type:          engine.ActionScout,
		ScoutFromLeft: m.scoutFromLeft,
		ScoutFlip:     m.scoutFlip,
		ScoutInsertAt: m.cursor,
	}

	if err := m.engine.ApplyAction(action); err != nil {
		m.statusMsg = "Invalid: " + err.Error()
		m.statusIsError = true
		return m, tickCmd(2000*1000000, ClearStatusMsg{})
	}

	m.statusMsg = ""
	return m.afterAction()
}

func (m GameModel) afterAction() (GameModel, tea.Cmd) {
	switch m.engine.Phase {
	case engine.PhaseRoundEnd:
		return m, func() tea.Msg { return RoundEndMsg{} }
	case engine.PhaseGameEnd:
		return m, func() tea.Msg { return GameOverMsg{} }
	}

	// Next player
	if m.engine.CurrentPlayer != m.humanID {
		m.uiPhase = UIPhaseWaitingForAI
		return m, aiDelayCmd(m.engine.CurrentPlayer)
	}

	if m.engine.Table.Combo == nil {
		m = m.enterShowSelect()
		m.statusMsg = "Table is empty — you must Show."
	} else {
		m.uiPhase = UIPhaseChooseAction
		m.cursor = 0
	}
	return m, nil
}

func (m GameModel) handleAITurn(msg AITurnMsg) (GameModel, tea.Cmd) {
	pid := msg.PlayerID
	if pid != m.engine.CurrentPlayer {
		return m, nil
	}

	bot := m.bots[pid]
	action := bot.ChooseAction(m.engine, pid)

	playerName := m.engine.Players[pid].Name
	m.statusMsg = fmt.Sprintf("%s chose %s", playerName, action.Type.String())

	if err := m.engine.ApplyAction(action); err != nil {
		m.statusMsg = fmt.Sprintf("%s error: %s", playerName, err.Error())
		m.statusIsError = true
		return m, nil
	}

	return m.afterAction()
}

func (m GameModel) View() string {
	var sections []string

	// Round info
	roundInfo := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render(
		fmt.Sprintf("═══ Round %d/%d ═══", m.engine.Round+1, m.engine.TotalRounds),
	)
	sections = append(sections, roundInfo)

	// Turn order indicator
	sections = append(sections, renderTurnOrder(m.engine))

	// Players table (all players including human)
	sections = append(sections, "")
	sections = append(sections, renderPlayersTable(m.engine))

	// Table
	sections = append(sections, "")
	ownerName := ""
	if m.engine.Table.Combo != nil {
		ownerName = m.engine.Players[m.engine.Table.OwnerID].Name
	}
	sections = append(sections, renderTableCombo(m.engine.Table, ownerName))

	// Player hand
	sections = append(sections, "")
	sections = append(sections, lipgloss.NewStyle().Foreground(colorGold).Bold(true).Render("Your Hand"))

	hand := m.engine.Players[m.humanID].Hand
	if m.uiPhase == UIPhaseScoutInsert {
		sections = append(sections, renderHandWithInsert(hand, m.cursor, m.scoutFromLeft, m.scoutFlip, m.engine.Table))
	} else {
		sections = append(sections, renderHand(hand, m.cursor, m.selected))
	}

	// Status / Action bar
	sections = append(sections, "")
	if m.statusMsg != "" {
		style := statusStyle
		if m.statusIsError {
			style = errorStyle
		}
		sections = append(sections, style.Render("  "+m.statusMsg))
	}

	sections = append(sections, m.viewActionHints())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m GameModel) viewActionHints() string {
	switch m.uiPhase {
	case UIPhaseFlipChoice:
		return fmt.Sprintf("  %s flip preview  %s confirm current orientation",
			keyStyle.Render("[F]"), keyStyle.Render("[Enter]"))
	case UIPhaseChooseAction:
		hints := fmt.Sprintf("  %s Show  %s sCout",
			keyStyle.Render("[1/S]"), keyStyle.Render("[2/C]"))
		if !m.engine.Players[m.humanID].UsedScoutShow {
			hints += fmt.Sprintf("  %s Scout&Show", keyStyle.Render("[3/B]"))
		}
		return hints
	case UIPhaseShowSelect:
		return fmt.Sprintf("  %s move  %s select  %s play  %s back",
			keyStyle.Render("←→"), keyStyle.Render("Space"), keyStyle.Render("Enter"), keyStyle.Render("Esc"))
	case UIPhaseScoutEnd:
		return fmt.Sprintf("  %s left end  %s right end  %s back",
			keyStyle.Render("←"), keyStyle.Render("→"), keyStyle.Render("Esc"))
	case UIPhaseScoutFlip:
		return fmt.Sprintf("  %s flip  %s keep  %s back",
			keyStyle.Render("[F]"), keyStyle.Render("[K/Enter]"), keyStyle.Render("Esc"))
	case UIPhaseScoutInsert:
		return fmt.Sprintf("  %s move position  %s insert  %s back",
			keyStyle.Render("←→"), keyStyle.Render("Enter"), keyStyle.Render("Esc"))
	case UIPhaseWaitingForAI:
		return descStyle.Render("  Waiting for opponent...")
	}
	return ""
}
