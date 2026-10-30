package feral

import (
	"strings"
	"testing"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/core/stats"
)

func whitemaneFeralRequest(seed int64, inFront bool, debug bool) *proto.RaidSimRequest {
	return &proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:      "Feral",
				Race:      proto.Race_RaceTauren,
				Class:     proto.Class_ClassDruid,
				Equipment: core.GetGearSet("../../../ui/feral_druid/gear_sets", "p1_feral").GearSet,
				Consumes:  FullConsumes,
				Spec:      PlayerOptionsMonoCatNoBleed,
				Buffs:     core.FullIndividualBuffs,
				Glyphs:    StandardGlyphs,
				Rotation:  core.GetAplRotation("../../../ui/feral_druid/apls", "meow_feral").Rotation,

				TalentsString:   StandardTalents,
				InFrontOfTarget: inFront,
			},
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			core.FullDebuffs),
		Encounter:  &proto.Encounter{Duration: 240, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: seed, Debug: debug, IsTest: true},
	}
}

func TestWhitemaneClearcastingBiteSpendsExtraEnergy(t *testing.T) {
	damage := map[float64]float64{}
	for _, energy := range []float64{0, 10, 30, 50} {
		sim := core.NewSim(whitemaneFeralRequest(1, false, false))
		sim.Reset()
		cat := sim.Raid.Parties[0].Players[0].(*FeralDruid)
		cat.AddStatsDynamic(sim, stats.Stats{
			stats.MeleeHit:  20 * core.MeleeHitRatingPerHitChance,
			stats.Expertise: 30 * 4 * core.ExpertisePerQuarterPercentReduction,
		})
		cat.AddComboPoints(sim, 5, cat.FerociousBite.ComboPointMetrics())
		cat.SpendEnergy(sim, cat.CurrentEnergy()-energy, cat.FerociousBite.Cost.(*core.EnergyCost).ResourceMetrics)
		cat.ClearcastingAura.Activate(sim)
		if !cat.FerociousBite.Cast(sim, cat.CurrentTarget) {
			t.Fatalf("energy %.0f: the Clearcasting Bite was not cast", energy)
		}
		if cat.FerociousBite.CurCast.Cost != 0 {
			t.Fatalf("energy %.0f: Clearcasting Bite cost %.3f, want 0", energy, cat.FerociousBite.CurCast.Cost)
		}
		if hits := cat.FerociousBite.SpellMetrics[0].Hits + cat.FerociousBite.SpellMetrics[0].Crits; hits != 1 {
			t.Fatalf("energy %.0f: the Bite did not land (%d hits)", energy, hits)
		}
		if got, want := cat.CurrentEnergy(), energy-min(energy, 30); got != want {
			t.Errorf("energy %.0f: %.3f energy left after a Clearcasting Bite, want %.3f (it spends up to 30 extra energy)", energy, got, want)
		}
		damage[energy] = cat.FerociousBite.SpellMetrics[0].TotalDamage
	}
	if !(damage[0] < damage[10] && damage[10] < damage[30] && damage[30] == damage[50]) {
		t.Errorf("Clearcasting Bite damage at 0, 10, 30 and 50 energy: %.1f, %.1f, %.1f, %.1f; want it to rise with the extra energy up to 30", damage[0], damage[10], damage[30], damage[50])
	}
}

func TestWhitemaneCatRotationNoClearcastingBite(t *testing.T) {
	free, paid := 0, 0
	for seed := int64(11); seed <= 20; seed++ {
		result := core.RunRaidSim(whitemaneFeralRequest(seed, true, true))
		if result.ErrorResult != "" {
			t.Fatal(result.ErrorResult)
		}
		for _, line := range strings.Split(result.Logs, "\n") {
			if !strings.Contains(line, "] [Feral (#1)] ") || !strings.Contains(line, "Casting {SpellID: 48577}") {
				continue
			}
			if strings.Contains(line, "(Cost = 0.000,") {
				free++
			} else {
				paid++
			}
		}
	}
	if paid < 20 {
		t.Fatalf("parser not calibrated: %d paid Bites in 10 fights of 240 s", paid)
	}
	if free > 0 {
		t.Fatalf("%d of %d Bites were cast on Clearcasting; a Clearcasting Bite only saves the base cost, so the rotation spends Clearcasting elsewhere", free, free+paid)
	}
	t.Logf("%d Bites, none on Clearcasting", paid)
}
