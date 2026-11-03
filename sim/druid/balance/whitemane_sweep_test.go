package balance

import (
	"math"
	"strings"
	"testing"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func TestWhitemaneStarfireWrathHighestRanks(t *testing.T) {
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
		Encounter:  &proto.Encounter{Duration: 180, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1, IsTest: true},
	})
	moonkin := sim.Raid.Parties[0].Players[0].(*BalanceDruid)
	if moonkin.Ranged().ID != 40321 {
		t.Fatalf("ranged slot holds %d, the p1 set should hold Idol of the Shooting Star (40321, +165 Starfire spell power)", moonkin.Ranged().ID)
	}
	wrathOfCenarius := float64(moonkin.Talents.WrathOfCenarius)

	type check struct {
		name           string
		lo, hi, oldMax float64
		power          func(sp float64) float64
		residuals      []float64
	}
	checks := map[core.ActionID]*check{
		moonkin.Starfire.ActionID: {name: "Starfire rank 10", lo: 1163, hi: 1369, oldMax: 1222, power: func(sp float64) float64 {
			return (sp+165)*1.12 + sp*0.04*wrathOfCenarius
		}},
		moonkin.Wrath.ActionID: {name: "Wrath rank 12", lo: 624, hi: 702, oldMax: 627, power: func(sp float64) float64 {
			return sp * (0.639968 + 0.02*wrathOfCenarius)
		}},
	}
	sim.Log = func(message string, vals ...interface{}) {
		if !strings.Contains(message, "[DEBUG]") || !strings.Contains(message, "BaseDamage") || len(vals) < 6 {
			return
		}
		id, ok := vals[1].(core.ActionID)
		if !ok {
			return
		}
		c := checks[id]
		if c == nil {
			return
		}
		sp, base := vals[4].(float64), vals[5].(float64)
		c.residuals = append(c.residuals, base-c.power(sp))
	}

	sim.Reset()
	sim.PrePull()
	for !sim.Step() {
	}
	sim.Cleanup()

	for _, c := range checks {
		if len(c.residuals) < 10 {
			t.Fatalf("%s: %d casts logged, want at least 10", c.name, len(c.residuals))
		}
		lo, hi := math.Inf(1), math.Inf(-1)
		for _, r := range c.residuals {
			lo, hi = math.Min(lo, r), math.Max(hi, r)
		}
		if lo < c.lo-1e-6 || hi > c.hi+1e-6 {
			t.Errorf("%s: base damage after the spell power part spans %.2f to %.2f over %d casts, want inside %.0f to %.0f", c.name, lo, hi, len(c.residuals), c.lo, c.hi)
		}
		if hi <= c.oldMax || lo > c.lo+(c.hi-c.lo)/4 {
			t.Errorf("%s: base damage spans %.2f to %.2f over %d casts, want it to reach above the stock maximum %.0f and into the lowest quarter above %.0f", c.name, lo, hi, len(c.residuals), c.oldMax, c.lo)
		}
	}
}
