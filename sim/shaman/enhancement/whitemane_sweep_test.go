package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/core/stats"
)

func whitemaneSweepFireElementalStats(race proto.Race, heroicPresence bool) stats.Stats {
	rsr := whitemaneRequest(1)
	rsr.Raid.Parties[0].Players[0].Race = race
	rsr.Raid.Parties[0].Buffs = &proto.PartyBuffs{HeroicPresence: heroicPresence}
	sim := core.NewSim(rsr)
	sim.Reset()
	enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
	enh.FireElemental.Enable(sim, enh.FireElemental)
	return enh.FireElemental.GetStats()
}

func TestWhitemaneSweepHeroicPresenceFireElemental(t *testing.T) {
	none := whitemaneSweepFireElementalStats(proto.Race_RaceOrc, false)
	for _, c := range []struct {
		name           string
		race           proto.Race
		heroicPresence bool
	}{
		{"Orc shaman in a party with Heroic Presence", proto.Race_RaceOrc, true},
		{"Draenei shaman", proto.Race_RaceDraenei, false},
	} {
		got := whitemaneSweepFireElementalStats(c.race, c.heroicPresence)
		melee := (got[stats.MeleeHit] - none[stats.MeleeHit]) / core.MeleeHitRatingPerHitChance
		spell := (got[stats.SpellHit] - none[stats.SpellHit]) / core.SpellHitRatingPerHitChance
		expertise := got[stats.Expertise] - none[stats.Expertise]
		t.Logf("%s: Fire Elemental %+.2f%% melee hit, %+.2f%% spell hit, %+.2f expertise rating against no Heroic Presence", c.name, melee, spell, expertise)
		if math.Abs(melee-1) > 1e-9 || math.Abs(spell-1) > 1e-9 || expertise != 0 {
			t.Errorf("%s: Heroic Presence gives the Fire Elemental %.2f%% melee hit, %.2f%% spell hit and %.2f expertise rating, want 1%%, 1%% and 0", c.name, melee, spell, expertise)
		}
	}
}
