package balance

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/core/stats"
	"github.com/wowsims/wotlk/sim/druid"
)

func TestWhitemaneTreants(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Moonkin",
				Race:          proto.Race_RaceTauren,
				Class:         proto.Class_ClassDruid,
				Equipment:     core.GetGearSet("../../../ui/balance_druid/gear_sets", "p1").GearSet,
				Consumes:      FullConsumes,
				Spec:          PlayerOptionsAdaptive,
				Buffs:         &proto.IndividualBuffs{},
				Glyphs:        StandardGlyphs,
				TalentsString: StandardTalents,
				Rotation:      core.GetAplRotation("../../../ui/balance_druid/apls", "basic_p3").Rotation,
			},
			&proto.PartyBuffs{},
			&proto.RaidBuffs{},
			&proto.Debuffs{}),
		Encounter:  &proto.Encounter{Duration: 180, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1, IsTest: true},
	})
	sim.Reset()
	moonkin := sim.Raid.Parties[0].Players[0].(*BalanceDruid)
	spellPower := moonkin.GetStat(stats.SpellPower)
	if spellPower < 1000 {
		t.Fatalf("owner spell power %.1f; the p1 set should give well over 1000", spellPower)
	}
	for i, treant := range []*druid.TreantPet{moonkin.Treant1, moonkin.Treant2, moonkin.Treant3} {
		if treant == nil {
			t.Fatalf("treant %d missing; Force of Nature is not talented", i+1)
		}
		if got := treant.AutoAttacks.MH().SwingSpeed; got != 1.5 {
			t.Errorf("treant %d swing speed %.2f s, want 1.5 s", i+1, got)
		}
		before := treant.GetStat(stats.AttackPower)
		treant.EnableWithTimeout(sim, treant, 30*time.Second)
		if got, want := treant.GetStat(stats.AttackPower)-before, 0.75*spellPower; math.Abs(got-want) > 1e-6 {
			t.Errorf("treant %d gains %.3f attack power from %.3f spell power, want %.3f (75%%)", i+1, got, spellPower, want)
		}
	}
}

func TestWhitemaneJudgementOfWisdomBalanceSpells(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Moonkin",
				Race:          proto.Race_RaceTauren,
				Class:         proto.Class_ClassDruid,
				Equipment:     core.GetGearSet("../../../ui/balance_druid/gear_sets", "p1").GearSet,
				Consumes:      FullConsumes,
				Spec:          PlayerOptionsAdaptive,
				Buffs:         core.FullIndividualBuffs,
				Glyphs:        StandardGlyphs,
				TalentsString: StandardTalents,
				Rotation:      core.GetAplRotation("../../../ui/balance_druid/apls", "basic_p3").Rotation,
			},
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			core.FullDebuffs),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1},
	})
	sim.Reset()
	moonkin := sim.Raid.Parties[0].Players[0].(*BalanceDruid)
	for _, c := range []struct {
		name  string
		spell *core.Spell
		procs bool
	}{
		{"Starfall star", moonkin.GetSpell(core.ActionID{SpellID: 53195}), false},
		{"Starfall splash", moonkin.GetSpell(core.ActionID{SpellID: 53190}), false},
		{"Wrath", moonkin.Wrath.Spell, true},
		{"Moonfire", moonkin.Moonfire.Spell, true},
	} {
		if c.spell == nil {
			t.Fatalf("%s not registered", c.name)
		}
		if got := core.JudgementOfWisdomManaEvents(sim, c.spell, 200); (got > 0) != c.procs {
			t.Errorf("%s: %d Judgement of Wisdom procs from 200 hits, want procs %v", c.name, got, c.procs)
		}
	}
}
