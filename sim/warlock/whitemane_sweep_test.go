package warlock

import (
	"math"
	"testing"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func whitemaneSweepDestroWithShadowTrance(t *testing.T) (*core.Simulation, *Warlock) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Lock",
				Race:          proto.Race_RaceOrc,
				Class:         proto.Class_ClassWarlock,
				Equipment:     core.GetGearSet("../../ui/warlock/gear_sets", "p4_destro").GearSet,
				TalentsString: DestructionTalents,
				Glyphs: &proto.Glyphs{
					Major1: int32(proto.WarlockMajorGlyph_GlyphOfConflagrate),
					Major2: int32(proto.WarlockMajorGlyph_GlyphOfLifeTap),
					Major3: int32(proto.WarlockMajorGlyph_GlyphOfCorruption),
				},
				Consumes: FullConsumes,
				Spec:     DefaultDestroWarlock,
				Buffs:    core.FullIndividualBuffs,
				Rotation: &proto.APLRotation{},
			},
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			core.FullDebuffs),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1},
	})
	sim.Reset()
	warlock := sim.Raid.Parties[0].Players[0].(*Warlock)
	if warlock.NightfallProcAura == nil || warlock.BackdraftAura == nil || warlock.Talents.Backdraft != 3 {
		t.Fatalf("setup without Shadow Trance or 3/3 Backdraft")
	}
	return sim, warlock
}

func TestWhitemaneSweepShadowTranceWithBackdraft(t *testing.T) {
	for _, backdraft := range []bool{false, true} {
		sim, warlock := whitemaneSweepDestroWithShadowTrance(t)
		if backdraft {
			warlock.BackdraftAura.Activate(sim)
			warlock.BackdraftAura.SetStacks(sim, 3)
		}
		warlock.NightfallProcAura.Activate(sim)
		if !warlock.ShadowBolt.Cast(sim, sim.Encounter.TargetUnits[0]) {
			t.Fatalf("Backdraft %v: Shadow Bolt not cast", backdraft)
		}
		castTime := warlock.ShadowBolt.CurCast.CastTime
		next := warlock.ApplyCastSpeedForSpell(warlock.ShadowBolt.DefaultCast.CastTime, warlock.ShadowBolt)
		want := warlock.ApplyCastSpeed(warlock.ShadowBolt.DefaultCast.CastTime)
		wantStacks := int32(0)
		if backdraft {
			want = warlock.ApplyCastSpeed(warlock.ShadowBolt.DefaultCast.CastTime * 7 / 10)
			wantStacks = 2
		}
		t.Logf("Backdraft %v: Shadow Trance cast time %s, Shadow Trance active %v, Backdraft stacks %d, next Shadow Bolt %s (want %s)",
			backdraft, castTime, warlock.NightfallProcAura.IsActive(), warlock.BackdraftAura.GetStacks(), next, want)
		if castTime > 0 {
			t.Errorf("Backdraft %v: the Shadow Trance Shadow Bolt took %s, want instant", backdraft, castTime)
		}
		if warlock.NightfallProcAura.IsActive() {
			t.Errorf("Backdraft %v: Shadow Trance still active after the instant Shadow Bolt", backdraft)
		}
		if math.Abs(float64(next-want)) > 1000 {
			t.Errorf("Backdraft %v: the next Shadow Bolt takes %s, want %s", backdraft, next, want)
		}
		if got := warlock.BackdraftAura.GetStacks(); got != wantStacks {
			t.Errorf("Backdraft %v: %d Backdraft charges left, want %d", backdraft, got, wantStacks)
		}
	}
}

func TestWhitemaneSweepShadowTranceKeptByHardcast(t *testing.T) {
	sim, warlock := whitemaneSweepDestroWithShadowTrance(t)
	target := sim.Encounter.TargetUnits[0]
	if !warlock.ShadowBolt.Cast(sim, target) || warlock.ShadowBolt.CurCast.CastTime <= 0 {
		t.Fatalf("no hard-cast Shadow Bolt started")
	}
	end := sim.CurrentTime + warlock.ShadowBolt.CurCast.CastTime
	warlock.NightfallProcAura.Activate(sim)
	for sim.CurrentTime <= end {
		if sim.Step() {
			t.Fatalf("the fight ended during the cast")
		}
	}
	casts := warlock.ShadowBolt.SpellMetrics[target.UnitIndex].Casts
	t.Logf("%d Shadow Bolt completed at %s, Shadow Trance active %v", casts, end, warlock.NightfallProcAura.IsActive())
	if casts != 1 {
		t.Fatalf("%d Shadow Bolts completed, want the one hard cast", casts)
	}
	if !warlock.NightfallProcAura.IsActive() {
		t.Errorf("the hard-cast Shadow Bolt consumed the Shadow Trance that came during its cast")
	}
}
