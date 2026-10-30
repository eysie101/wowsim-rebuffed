package retribution

import (
	"testing"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func TestWhitemaneJudgementOfWisdomRetributionSpells(t *testing.T) {
	equipment := core.GetGearSet("../../../ui/retribution_paladin/gear_sets", "p1").GearSet
	equipment.Items[14].Id = 130031
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Ret",
				Race:          proto.Race_RaceBloodElf,
				Class:         proto.Class_ClassPaladin,
				Equipment:     equipment,
				TalentsString: StandardTalents,
				Glyphs: &proto.Glyphs{
					Major1: int32(proto.PaladinMajorGlyph_GlyphOfSealOfVengeance),
					Major2: int32(proto.PaladinMajorGlyph_GlyphOfJudgement),
					Major3: int32(proto.PaladinMajorGlyph_GlyphOfReckoning),
				},
				Consumes: FullConsumes,
				Spec:     DefaultOptions,
				Buffs:    core.FullIndividualBuffs,
				Rotation: core.GetAplRotation("../../../ui/retribution_paladin/apls", "default").Rotation,
			},
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			core.FullDebuffs),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1},
	})
	sim.Reset()
	ret := sim.Raid.Parties[0].Players[0].(*RetributionPaladin)
	for _, c := range []struct {
		name  string
		spell *core.Spell
		procs bool
	}{
		{"Judgement of Command", ret.GetSpell(core.ActionID{SpellID: 20467}), false},
		{"Judgement of Righteousness", ret.GetSpell(core.ActionID{SpellID: 20187}), false},
		{"Judgement of Vengeance", ret.GetSpell(core.ActionID{SpellID: 31804}), false},
		{"Hand of Reckoning damage", ret.HandOfReckoning, false},
		{"Armata Strigoi attack", ret.GetSpell(core.ActionID{SpellID: 36735}), false},
		{"main-hand swing", ret.AutoAttacks.MHAuto(), true},
		{"Crusader Strike", ret.CrusaderStrike, true},
		{"Exorcism", ret.Exorcism, true},
	} {
		if c.spell == nil {
			t.Fatalf("%s not registered", c.name)
		}
		if got := core.JudgementOfWisdomManaEvents(sim, c.spell, 200); (got > 0) != c.procs {
			t.Errorf("%s: %d Judgement of Wisdom procs from 200 hits, want procs %v", c.name, got, c.procs)
		}
	}
}
