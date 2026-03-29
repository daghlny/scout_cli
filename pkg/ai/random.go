package ai

import (
	"math/rand"

	"github.com/daghlny/scout_cli/pkg/engine"
)

// RandomBot picks uniformly at random from all legal actions.
type RandomBot struct {
	rng *rand.Rand
}

func NewRandomBot(rng *rand.Rand) *RandomBot {
	return &RandomBot{rng: rng}
}

func (b *RandomBot) Name() string { return "Random" }

func (b *RandomBot) ChooseHandOrientation(hand *engine.Hand, state *engine.GameState) bool {
	return b.rng.Intn(2) == 1
}

func (b *RandomBot) ChooseAction(state *engine.GameState, playerID int) engine.Action {
	actions := engine.ValidActions(state)
	return actions[b.rng.Intn(len(actions))]
}
