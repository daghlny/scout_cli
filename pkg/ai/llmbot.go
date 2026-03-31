package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	"github.com/daghlny/scout_cli/pkg/engine"
	"github.com/daghlny/scout_cli/pkg/llm"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// LLMBot uses a language model to make strategic decisions.
type LLMBot struct {
	client       *llm.Client
	fallback     Strategy
	history      []llm.ChatMessage // thinking history for continuity
	roundHistory []string          // action log for current round
	lastRound    int
	LastError    string // exposed for TUI to display
}

func NewLLMBot(client *llm.Client, rng *rand.Rand) *LLMBot {
	return &LLMBot{
		client:   client,
		fallback: NewSmartBot(rng),
	}
}

func (b *LLMBot) Name() string { return "LLM" }

func (b *LLMBot) ChooseHandOrientation(hand *engine.Hand, state *engine.GameState) bool {
	// Use SmartBot for orientation — fast and reliable
	return b.fallback.ChooseHandOrientation(hand, state)
}

func (b *LLMBot) ChooseAction(state *engine.GameState, playerID int) engine.Action {
	b.LastError = ""

	// Reset history on new round
	if state.Round != b.lastRound {
		b.history = nil
		b.roundHistory = nil
		b.lastRound = state.Round
	}

	action, thinking, err := b.chooseActionLLM(state, playerID)
	if err != nil {
		b.LastError = err.Error()
		return b.fallback.ChooseAction(state, playerID)
	}

	// Record thinking for continuity
	if thinking != "" {
		b.history = append(b.history, llm.ChatMessage{
			Role:    "assistant",
			Content: fmt.Sprintf("[Turn thinking]: %s", thinking),
		})
		// Keep history manageable (last 10 entries)
		if len(b.history) > 10 {
			b.history = b.history[len(b.history)-10:]
		}
	}

	return action
}

// RecordAction records any player's action for context.
func (b *LLMBot) RecordAction(playerName string, action engine.Action, state *engine.GameState) {
	desc := describeAction(playerName, action, state)
	b.roundHistory = append(b.roundHistory, desc)
	// Keep manageable
	if len(b.roundHistory) > 30 {
		b.roundHistory = b.roundHistory[len(b.roundHistory)-30:]
	}
}

func (b *LLMBot) chooseActionLLM(state *engine.GameState, playerID int) (engine.Action, string, error) {
	validActions := engine.ValidActions(state)
	if len(validActions) == 0 {
		return engine.Action{}, "", fmt.Errorf("no valid actions")
	}

	systemPrompt := buildSystemPrompt()
	userPrompt := buildUserPrompt(state, playerID, b.roundHistory)

	// Build messages: system + history + current user prompt
	var messages []llm.ChatMessage
	messages = append(messages, llm.ChatMessage{Role: "system", Content: systemPrompt})
	messages = append(messages, b.history...)
	messages = append(messages, llm.ChatMessage{Role: "user", Content: userPrompt})

	tools := []llm.Tool{chooseActionTool()}

	resp, err := b.client.Chat(context.Background(), messages, tools)
	if err != nil {
		return engine.Action{}, "", err
	}

	// Parse tool call
	if len(resp.ToolCalls) == 0 {
		return engine.Action{}, "", fmt.Errorf("LLM did not call choose_action function")
	}

	tc := resp.ToolCalls[0]
	if tc.Name != "choose_action" {
		return engine.Action{}, "", fmt.Errorf("unexpected function: %s", tc.Name)
	}

	action, thinking, err := parseActionResponse(tc.Arguments, state, playerID, validActions)
	if err != nil {
		return engine.Action{}, thinking, fmt.Errorf("invalid LLM action: %w", err)
	}

	return action, thinking, nil
}

// === Prompt Building ===

func buildSystemPrompt() string {
	return `You are an expert Scout card game player. Scout uses 45 double-sided cards (numbers 1-10).

RULES:
- Cards in hand CANNOT be rearranged. Only insertion via Scout is allowed.
- Actions: Show (play adjacent cards beating table), Scout (take from table end, insert in hand), Scout&Show (both, once per round).
- Valid combos: single, set (same numbers), run (consecutive asc/desc) — must be adjacent in hand.
- Beating: more cards > fewer; same count: set > run; same type: higher min value wins.
- Round ends when someone empties hand OR all others scout consecutively (table owner wins exempt from hand penalty).
- Scoring: +1 per collected card, +1 per scout token, -1 per remaining hand card.

STRATEGY TIPS:
- Save strong combos for later; play "just enough" to beat the table.
- When hand is small (≤4), prioritize clearing cards.
- If ConsecutiveScouts is high and you're NOT table owner, you MUST show to prevent opponent winning.
- Scout insert position matters: place next to matching/consecutive values to build combos.
- Scout&Show is precious — save for high-impact plays (3+ cards).

Use the choose_action function to respond. Always include your thinking process.`
}

func buildUserPrompt(state *engine.GameState, playerID int, roundHistory []string) string {
	player := &state.Players[playerID]
	hand := player.Hand
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== Round %d/%d, Your Turn ===\n", state.Round+1, state.TotalRounds))
	sb.WriteString(fmt.Sprintf("You are Player %d (%s)\n\n", playerID, player.Name))

	// Your hand
	sb.WriteString("YOUR HAND (left to right, index 0 to N-1):\n")
	entries := hand.Entries()
	for i, e := range entries {
		sb.WriteString(fmt.Sprintf("  [%d] active=%d (other side=%d)\n", i, e.ActiveValue(), e.InactiveValue()))
	}
	sb.WriteString(fmt.Sprintf("  Total: %d cards\n\n", hand.Len()))

	// Table
	if state.Table.Combo != nil {
		owner := state.Players[state.Table.OwnerID].Name
		sb.WriteString(fmt.Sprintf("TABLE: %s (%d cards, min=%d) by %s\n",
			state.Table.Combo.Type.String(), state.Table.Combo.Size(),
			state.Table.Combo.MinValue(), owner))
		sb.WriteString("  Cards: ")
		for _, e := range state.Table.Entries {
			sb.WriteString(fmt.Sprintf("[%d] ", e.ActiveValue()))
		}
		sb.WriteString("\n\n")
	} else {
		sb.WriteString("TABLE: Empty — you must Show.\n\n")
	}

	// Players info
	sb.WriteString("ALL PLAYERS:\n")
	for _, p := range state.Players {
		marker := ""
		if p.ID == playerID {
			marker = " (YOU)"
		}
		sb.WriteString(fmt.Sprintf("  %s%s: %d cards, %d collected, %d tokens, S&S %s\n",
			p.Name, marker, p.Hand.Len(), len(p.CollectedCards), p.ScoutTokens,
			map[bool]string{true: "used", false: "available"}[p.UsedScoutShow]))
	}
	sb.WriteString(fmt.Sprintf("\nConsecutiveScouts: %d (round ends at %d)\n", state.ConsecutiveScouts, state.NumPlayers-1))

	if state.Table.Combo != nil {
		isOwner := state.Table.OwnerID == playerID
		if state.ConsecutiveScouts >= state.NumPlayers-2 && !isOwner {
			sb.WriteString("⚠️ WARNING: Next scout ends the round for opponent! You SHOULD show if possible!\n")
		}
	}

	// Round history
	if len(roundHistory) > 0 {
		sb.WriteString("\nRECENT ACTIONS THIS ROUND:\n")
		start := 0
		if len(roundHistory) > 10 {
			start = len(roundHistory) - 10
		}
		for _, h := range roundHistory[start:] {
			sb.WriteString(fmt.Sprintf("  %s\n", h))
		}
	}

	// Available action summary
	sb.WriteString("\nAVAILABLE ACTIONS:\n")
	hasShow, hasScout, hasSS := false, false, false
	for _, a := range engine.ValidActions(state) {
		switch a.Type {
		case engine.ActionShow:
			hasShow = true
		case engine.ActionScout:
			hasScout = true
		case engine.ActionScoutAndShow:
			hasSS = true
		}
	}
	if hasShow {
		sb.WriteString("  - SHOW: Play adjacent cards from hand that beat the table\n")
	}
	if hasScout {
		sb.WriteString("  - SCOUT: Take a card from table's left or right end\n")
	}
	if hasSS {
		sb.WriteString("  - SCOUT_AND_SHOW: Scout then show (once per round)\n")
	}

	sb.WriteString("\nChoose your action using the choose_action function.")
	return sb.String()
}

// === Tool Definition ===

func chooseActionTool() llm.Tool {
	return llm.Tool{
		Name:        "choose_action",
		Description: "Choose an action for your turn in the Scout card game",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"thinking": {
					Type:        jsonschema.String,
					Description: "Your analysis and reasoning for this decision (in any language)",
				},
				"action_type": {
					Type:        jsonschema.String,
					Enum:        []string{"show", "scout", "scout_and_show"},
					Description: "The type of action to take",
				},
				"show_card_indices": {
					Type:        jsonschema.Array,
					Description: "Indices of adjacent hand cards to play (for show/scout_and_show). Must be consecutive indices.",
					Items: &jsonschema.Definition{
						Type: jsonschema.Integer,
					},
				},
				"scout_from_left": {
					Type:        jsonschema.Boolean,
					Description: "Take card from left end (true) or right end (false) of table combo",
				},
				"scout_flip": {
					Type:        jsonschema.Boolean,
					Description: "Whether to flip the scouted card before inserting",
				},
				"scout_insert_at": {
					Type:        jsonschema.Integer,
					Description: "Position in hand to insert scouted card (0 = before first, hand_size = after last)",
				},
			},
			Required: []string{"thinking", "action_type"},
		},
	}
}

// === Response Parsing ===

type actionResponse struct {
	Thinking        string `json:"thinking"`
	ActionType      string `json:"action_type"`
	ShowCardIndices []int  `json:"show_card_indices"`
	ScoutFromLeft   *bool  `json:"scout_from_left"`
	ScoutFlip       *bool  `json:"scout_flip"`
	ScoutInsertAt   *int   `json:"scout_insert_at"`
}

func parseActionResponse(argsJSON string, state *engine.GameState, playerID int, validActions []engine.Action) (engine.Action, string, error) {
	var resp actionResponse
	if err := json.Unmarshal([]byte(argsJSON), &resp); err != nil {
		return engine.Action{}, "", fmt.Errorf("JSON parse error: %w", err)
	}

	switch resp.ActionType {
	case "show":
		return matchShowAction(resp, state, playerID, validActions)
	case "scout":
		return matchScoutAction(resp, state, playerID, validActions)
	case "scout_and_show":
		return matchScoutAndShowAction(resp, state, playerID, validActions)
	default:
		return engine.Action{}, resp.Thinking, fmt.Errorf("unknown action_type: %s", resp.ActionType)
	}
}

func matchShowAction(resp actionResponse, state *engine.GameState, playerID int, validActions []engine.Action) (engine.Action, string, error) {
	if len(resp.ShowCardIndices) == 0 {
		return engine.Action{}, resp.Thinking, fmt.Errorf("show requires card indices")
	}

	start := resp.ShowCardIndices[0]
	count := len(resp.ShowCardIndices)

	// Find matching valid action
	for _, a := range validActions {
		if a.Type == engine.ActionShow && a.ShowStart == start && a.ShowCount == count {
			return a, resp.Thinking, nil
		}
	}

	return engine.Action{}, resp.Thinking, fmt.Errorf("no valid show action for indices %v", resp.ShowCardIndices)
}

func matchScoutAction(resp actionResponse, state *engine.GameState, playerID int, validActions []engine.Action) (engine.Action, string, error) {
	fromLeft := true
	if resp.ScoutFromLeft != nil {
		fromLeft = *resp.ScoutFromLeft
	}
	flip := false
	if resp.ScoutFlip != nil {
		flip = *resp.ScoutFlip
	}
	insertAt := 0
	if resp.ScoutInsertAt != nil {
		insertAt = *resp.ScoutInsertAt
	}

	// Find matching valid action
	for _, a := range validActions {
		if a.Type == engine.ActionScout &&
			a.ScoutFromLeft == fromLeft &&
			a.ScoutFlip == flip &&
			a.ScoutInsertAt == insertAt {
			return a, resp.Thinking, nil
		}
	}

	// Try to find any scout with same direction as fallback
	for _, a := range validActions {
		if a.Type == engine.ActionScout && a.ScoutFromLeft == fromLeft && a.ScoutFlip == flip {
			return a, resp.Thinking, nil
		}
	}

	return engine.Action{}, resp.Thinking, fmt.Errorf("no valid scout action")
}

func matchScoutAndShowAction(resp actionResponse, state *engine.GameState, playerID int, validActions []engine.Action) (engine.Action, string, error) {
	fromLeft := true
	if resp.ScoutFromLeft != nil {
		fromLeft = *resp.ScoutFromLeft
	}
	flip := false
	if resp.ScoutFlip != nil {
		flip = *resp.ScoutFlip
	}
	insertAt := 0
	if resp.ScoutInsertAt != nil {
		insertAt = *resp.ScoutInsertAt
	}

	start := 0
	count := 1
	if len(resp.ShowCardIndices) > 0 {
		start = resp.ShowCardIndices[0]
		count = len(resp.ShowCardIndices)
	}

	// Find matching valid action
	for _, a := range validActions {
		if a.Type == engine.ActionScoutAndShow &&
			a.ScoutFromLeft == fromLeft &&
			a.ScoutFlip == flip &&
			a.ScoutInsertAt == insertAt &&
			a.ShowStart == start &&
			a.ShowCount == count {
			return a, resp.Thinking, nil
		}
	}

	return engine.Action{}, resp.Thinking, fmt.Errorf("no valid scout_and_show action")
}

// === Helpers ===

func describeAction(playerName string, action engine.Action, state *engine.GameState) string {
	switch action.Type {
	case engine.ActionShow:
		if state.Table.Combo != nil {
			return fmt.Sprintf("%s showed %s (%d cards)", playerName, state.Table.Combo.Type.String(), state.Table.Combo.Size())
		}
		return fmt.Sprintf("%s showed cards", playerName)
	case engine.ActionScout:
		side := "left"
		if !action.ScoutFromLeft {
			side = "right"
		}
		return fmt.Sprintf("%s scouted from %s", playerName, side)
	case engine.ActionScoutAndShow:
		return fmt.Sprintf("%s used Scout & Show", playerName)
	}
	return fmt.Sprintf("%s took an action", playerName)
}
