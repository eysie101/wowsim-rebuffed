package shadow

import (
	"testing"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/core/stats"
)

func TestWhitemaneShadowfiendKeepsOwnerShareAfterReset(t *testing.T) {
	rsr := &proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Priest",
				Race:          proto.Race_RaceUndead,
				Class:         proto.Class_ClassPriest,
				Equipment:     core.GetGearSet("../../../ui/shadow_priest/gear_sets", "p1").GearSet,
				TalentsString: DefaultTalents,
				Glyphs:        DefaultGlyphs,
				Consumes:      FullConsumes,
				Spec:          PlayerOptionsIdeal,
				Buffs:         core.FullIndividualBuffs,
				Rotation:      core.GetAplRotation("../../../ui/shadow_priest/apls", "default").Rotation,
			},
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			core.FullDebuffs),
		Encounter:  &proto.Encounter{Duration: 10, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1},
	}
	sim := core.NewSim(rsr)
	priest := sim.Raid.Parties[0].Players[0].(*ShadowPriest)
	fiend := priest.ShadowfiendPet
	var base stats.Stats
	var firstGain float64
	for iteration := 1; iteration <= 4; iteration++ {
		sim.Reset()
		got := fiend.GetStats()
		if iteration == 1 {
			base = got
		} else if !got.Equals(base) {
			t.Errorf("iteration %d: the Shadowfiend starts with %.1f attack power, %.1f melee hit and %.1f expertise rating; iteration 1 started with %.1f, %.1f and %.1f", iteration, got[stats.AttackPower], got[stats.MeleeHit], got[stats.Expertise], base[stats.AttackPower], base[stats.MeleeHit], base[stats.Expertise])
		}
		sim.PrePull()
		for !sim.Step() {
		}
		if !fiend.IsEnabled() {
			t.Fatalf("iteration %d: the Shadowfiend is not out when the fight ends", iteration)
		}
		gain := fiend.GetStat(stats.AttackPower) - base[stats.AttackPower]
		if iteration == 1 {
			firstGain = gain
			if gain < 1000 {
				t.Fatalf("iteration 1: the Shadowfiend inherited %.1f attack power, want a large share of the owner's spell power", gain)
			}
		} else if gain < 0.9*firstGain {
			t.Errorf("iteration %d: the Shadowfiend inherited %.1f attack power by the end, iteration 1 %.1f", iteration, gain, firstGain)
		}
		sim.Cleanup()
	}
}
