package shadow

import (
	"testing"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func TestWhitemaneJudgementOfWisdomShadowSpells(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Shadow",
				Race:          proto.Race_RaceUndead,
				Class:         proto.Class_ClassPriest,
				Equipment:     core.GetGearSet("../../../ui/shadow_priest/gear_sets", "p1").GearSet,
				TalentsString: DefaultTalents,
				Glyphs:        DefaultGlyphs,
				Consumes:      FullConsumes,
				Spec:          PlayerOptionsBasic,
				Buffs:         core.FullIndividualBuffs,
				Rotation:      core.GetAplRotation("../../../ui/shadow_priest/apls", "default").Rotation,
			},
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			core.FullDebuffs),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1},
	})
	sim.Reset()
	priest := sim.Raid.Parties[0].Players[0].(*ShadowPriest)
	for _, c := range []struct {
		name  string
		spell *core.Spell
		procs bool
	}{
		{"Improved Devouring Plague", priest.GetSpell(core.ActionID{SpellID: 63675}), false},
		{"Devouring Plague cast", priest.DevouringPlague, true},
		{"Mind Blast", priest.MindBlast, true},
	} {
		if c.spell == nil {
			t.Fatalf("%s not registered", c.name)
		}
		if got := core.JudgementOfWisdomManaEvents(sim, c.spell, 200); (got > 0) != c.procs {
			t.Errorf("%s: %d Judgement of Wisdom procs from 200 hits, want procs %v", c.name, got, c.procs)
		}
	}
}
