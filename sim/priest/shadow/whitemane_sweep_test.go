package shadow

import (
	"testing"
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func TestWhitemaneSweepGlyphOfQuickPain(t *testing.T) {
	for _, c := range []struct {
		name   string
		major3 proto.PriestMajorGlyph
		hasted bool
	}{
		{"Glyph of Dispersion", proto.PriestMajorGlyph_GlyphOfDispersion, false},
		{"Glyph of Quick Pain", proto.PriestMajorGlyph_GlyphOfQuickPain, true},
	} {
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid: core.SinglePlayerRaidProto(
				&proto.Player{
					Name:          "Shadow",
					Race:          proto.Race_RaceUndead,
					Class:         proto.Class_ClassPriest,
					Equipment:     core.GetGearSet("../../../ui/shadow_priest/gear_sets", "p1").GearSet,
					TalentsString: DefaultTalents,
					Glyphs: &proto.Glyphs{
						Major1: int32(proto.PriestMajorGlyph_GlyphOfShadow),
						Major2: int32(proto.PriestMajorGlyph_GlyphOfMindFlay),
						Major3: int32(c.major3),
					},
					Consumes: FullConsumes,
					Spec:     PlayerOptionsBasic,
					Buffs:    core.FullIndividualBuffs,
					Rotation: core.GetAplRotation("../../../ui/shadow_priest/apls", "default").Rotation,
				},
				core.FullPartyBuffs,
				core.FullRaidBuffs,
				core.FullDebuffs),
			Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
			SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1},
		})
		sim.Reset()
		priest := sim.Raid.Parties[0].Players[0].(*ShadowPriest)
		target := sim.Encounter.TargetUnits[0]
		dot := priest.ShadowWordPain.Dot(target)
		dot.Apply(sim)

		hastedPeriod := priest.ApplyCastSpeedForSpell(3*time.Second, priest.ShadowWordPain)
		if hastedPeriod >= 3*time.Second {
			t.Fatalf("%s: the priest's spell haste gives a %s period, want below 3s for this test", c.name, hastedPeriod)
		}
		want := 3 * time.Second
		if c.hasted {
			want = hastedPeriod
		}
		if got := dot.TickPeriod(); got != want {
			t.Errorf("%s: Shadow Word: Pain ticks every %s, want %s", c.name, got, want)
		}
		if dot.NumberOfTicks != 6 {
			t.Errorf("%s: Shadow Word: Pain has %d ticks, want 6", c.name, dot.NumberOfTicks)
		}
		if got := dot.Aura.Duration; got != want*6 {
			t.Errorf("%s: Shadow Word: Pain lasts %s, want %s", c.name, got, want*6)
		}
	}
}
