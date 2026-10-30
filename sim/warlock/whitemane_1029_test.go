package warlock

import (
	"strings"
	"testing"
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func whitemanePetCastStarts(t *testing.T, gearSet string, talents string, glyphs *proto.Glyphs, options *proto.Player_Warlock, apl string, spellID int32) (time.Duration, []time.Duration) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Lock",
				Race:          proto.Race_RaceOrc,
				Class:         proto.Class_ClassWarlock,
				Equipment:     core.GetGearSet("../../ui/warlock/gear_sets", gearSet).GearSet,
				TalentsString: talents,
				Glyphs:        glyphs,
				Consumes:      FullConsumes,
				Spec:          options,
				Buffs:         core.FullIndividualBuffs,
				Rotation:      core.GetAplRotation("../../ui/warlock/apls", apl).Rotation,
			},
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			core.FullDebuffs),
		Encounter: &proto.Encounter{
			Duration: 180,
			Targets:  []*proto.Target{core.NewDefaultTarget()},
		},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1, IsTest: true},
	})

	pet := sim.Raid.Parties[0].Players[0].(*Warlock).Pet
	prefix := pet.LogLabel() + " "
	actionID := core.ActionID{SpellID: spellID}
	summoned := core.NeverExpires
	var starts []time.Duration
	sim.Log = func(message string, vals ...interface{}) {
		if !strings.HasPrefix(message, prefix) {
			return
		}
		if strings.HasSuffix(message, "Pet summoned") {
			summoned = sim.CurrentTime
		}
		if strings.HasPrefix(message, prefix+"Casting ") && len(vals) > 0 && vals[0] == actionID {
			starts = append(starts, sim.CurrentTime)
		}
	}

	sim.Reset()
	sim.PrePull()
	for !sim.Step() {
	}
	sim.Cleanup()

	if summoned == core.NeverExpires {
		t.Fatal("the pet was never summoned")
	}
	return summoned, starts
}

func whitemaneCheckGrid(t *testing.T, name string, summoned time.Duration, starts []time.Duration, minCasts int, minGap time.Duration) {
	t.Logf("%d casts of %s, pet summoned at %s, first starts %v", len(starts), name, summoned, starts[:min(6, len(starts))])
	if len(starts) < minCasts {
		t.Fatalf("%d casts of %s, want at least %d", len(starts), name, minCasts)
	}
	for i, start := range starts {
		if offset := (start - summoned) % (800 * time.Millisecond); offset != 0 {
			t.Errorf("%s %d starts at %s, %s after the last 0.8 s grid point", name, i+1, start, offset)
		}
		if i > 0 && start-starts[i-1] < minGap {
			t.Errorf("%s %d starts %s after the previous one, want at least %s", name, i+1, start-starts[i-1], minGap)
		}
	}
}

func TestWhitemaneImpFireboltGrid(t *testing.T) {
	summoned, starts := whitemanePetCastStarts(t, "p4_destro", DestructionTalents, DestructionGlyphs, DefaultDestroWarlock, "destro", 47964)
	whitemaneCheckGrid(t, "Firebolt", summoned, starts, 30, 800*time.Millisecond)
}

func TestWhitemaneFelguardCleaveGrid(t *testing.T) {
	summoned, starts := whitemanePetCastStarts(t, "p4_demo", DemonologyTalents, DemonologyGlyphs, DefaultDemonologyWarlock, "demo", 47994)
	whitemaneCheckGrid(t, "Cleave", summoned, starts, 20, 6*time.Second)
}

func TestWhitemaneJudgementOfWisdomWarlockSpells(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Lock",
				Race:          proto.Race_RaceOrc,
				Class:         proto.Class_ClassWarlock,
				Equipment:     core.GetGearSet("../../ui/warlock/gear_sets", "p4_affliction").GearSet,
				TalentsString: AfflictionTalents,
				Glyphs:        AfflictionGlyphs,
				Consumes:      FullConsumes,
				Spec:          DefaultAfflictionWarlock,
				Buffs:         core.FullIndividualBuffs,
				Rotation:      core.GetAplRotation("../../ui/warlock/apls", "affliction").Rotation,
			},
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			core.FullDebuffs),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1},
	})
	sim.Reset()
	warlock := sim.Raid.Parties[0].Players[0].(*Warlock)
	for _, c := range []struct {
		name  string
		spell *core.Spell
		procs bool
	}{
		{"Seed of Corruption explosion", warlock.GetSpell(core.ActionID{SpellID: 47836, Tag: 1}), false},
		{"Shadow Bolt", warlock.ShadowBolt, true},
		{"Haunt", warlock.Haunt, true},
	} {
		if c.spell == nil {
			t.Fatalf("%s not registered", c.name)
		}
		if got := core.JudgementOfWisdomManaEvents(sim, c.spell, 200); (got > 0) != c.procs {
			t.Errorf("%s: %d Judgement of Wisdom procs from 200 hits, want procs %v", c.name, got, c.procs)
		}
	}
}
