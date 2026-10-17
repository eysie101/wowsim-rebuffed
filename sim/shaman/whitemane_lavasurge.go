package shaman

import (
	"time"

	"github.com/wowsims/wotlk/sim/core"
)

// Doomhammer (Whitemane 2, items 131004 and 315008) carries Lava Surge (equip spell 9250001): each Flame Shock tick has a
// 20% chance to reset the cooldown of Lava Burst and to make the next one instant (buff 8251314, 3 s in the client's
// Spell.dbc).
var doomhammerIDs = map[int32]bool{131004: true, 315008: true}

const lavaSurgeChance = 0.20
const lavaSurgeDuration = 3 * time.Second

func (shaman *Shaman) registerLavaSurge() {
	if shaman.LavaBurst == nil || !doomhammerIDs[shaman.MainHand().ID] {
		return
	}
	lavaBurst := shaman.LavaBurst
	shaman.lavaSurgeAura = shaman.RegisterAura(core.Aura{
		Label:    "Lava Surge",
		ActionID: core.ActionID{SpellID: 8251314},
		Duration: lavaSurgeDuration,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			lavaBurst.CD.Reset()
			lavaBurst.CastTimeMultiplier -= 1
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			lavaBurst.CastTimeMultiplier += 1
		},
		OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell == lavaBurst {
				aura.Deactivate(sim)
			}
		},
	})
}

func (shaman *Shaman) lavaSurgeOnFlameShockTick(sim *core.Simulation) {
	if shaman.lavaSurgeAura != nil && sim.RandomFloat("Lava Surge") < lavaSurgeChance {
		shaman.lavaSurgeAura.Activate(sim)
	}
}
