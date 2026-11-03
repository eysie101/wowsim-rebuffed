package hunter

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func whitemaneStormcoilSim(rangedID int32, numTargets int, debuffs *proto.Debuffs) (*core.Simulation, *Hunter) {
	gear := core.GetGearSet("../../ui/hunter/gear_sets", "p1_mm").GearSet
	gear.Items[proto.ItemSlot_ItemSlotRanged].Id = rangedID
	encounter := &proto.Encounter{Duration: 60}
	for i := 0; i < numTargets; i++ {
		encounter.Targets = append(encounter.Targets, core.NewDefaultTarget())
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Hunter",
				Race:          proto.Race_RaceOrc,
				Class:         proto.Class_ClassHunter,
				Equipment:     gear,
				Consumes:      FullConsumes,
				TalentsString: MMTalents,
				Glyphs:        MMGlyphs,
				Spec:          PlayerOptionsBasic,
				Buffs:         core.FullIndividualBuffs,
				Rotation:      core.GetAplRotation("../../ui/hunter/apls", "mm").Rotation,
			},
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			debuffs),
		Encounter:  encounter,
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1},
	})
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*Hunter)
}

func TestWhitemaneSweepStormcoil(t *testing.T) {
	chainLightningID := core.ActionID{SpellID: 25442}
	for _, c := range []struct {
		id  int32
		has bool
	}{{150005, false}, {150090, true}, {315005, true}} {
		_, hunter := whitemaneStormcoilSim(c.id, 1, core.FullDebuffs)
		if got := hunter.GetSpell(chainLightningID) != nil; got != c.has {
			t.Errorf("Stormcoil %d: Chain Lightning registered %v, want %v", c.id, got, c.has)
		}
	}

	sim, hunter := whitemaneStormcoilSim(150090, 4, &proto.Debuffs{})
	chainLightning := hunter.GetSpell(chainLightningID)
	trigger := hunter.GetAura("Stormcoil Trigger")
	if chainLightning == nil || trigger == nil {
		t.Fatal("Stormcoil 150090 has no Chain Lightning proc")
	}
	target := sim.Encounter.TargetUnits[0]
	casts := func() int32 { return chainLightning.SpellMetrics[target.UnitIndex].Casts }
	hit := func(spell *core.Spell) bool {
		before := casts()
		trigger.OnSpellHitDealt(trigger, sim, spell, &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 1})
		return casts() > before
	}

	const rangedHits = 200_000
	for i := 0; i < rangedHits; i++ {
		sim.CurrentTime += 21 * time.Second
		hit(hunter.AutoAttacks.RangedAuto())
	}
	want := 2.075 * 2.9 / 60
	rate := float64(casts()) / rangedHits
	t.Logf("%d procs from %d landed Auto Shots with the internal cooldown over: %.4f per hit, want %.4f", casts(), rangedHits, rate, want)
	if math.Abs(rate-want) > 0.002 {
		t.Errorf("proc chance per landed ranged hit %.4f, want %.4f (2.075 per minute at 2.9 s)", rate, want)
	}

	meleeProcs := 0
	for i := 0; i < 2000; i++ {
		sim.CurrentTime += 21 * time.Second
		if hit(hunter.RaptorStrike) {
			meleeProcs++
		}
	}
	if meleeProcs != 0 {
		t.Errorf("%d procs from 2000 Raptor Strikes, want 0", meleeProcs)
	}

	var procs []time.Duration
	for i := 0; i < 3000; i++ {
		sim.CurrentTime += time.Second
		if hit(hunter.SteadyShot) {
			procs = append(procs, sim.CurrentTime)
		}
	}
	minGap := time.Hour
	for i := 1; i < len(procs); i++ {
		minGap = min(minGap, procs[i]-procs[i-1])
	}
	t.Logf("%d procs from 3000 Steady Shots one second apart, shortest gap %s", len(procs), minGap)
	if len(procs) < 50 || minGap != 20*time.Second {
		t.Errorf("%d procs, shortest gap %s; want at least 50 and a 20 s internal cooldown", len(procs), minGap)
	}

	m := chainLightning.SpellMetrics
	landed := func(i int) int32 { return m[i].Hits + m[i].Crits + m[i].Misses }
	for i, want := range []int32{casts(), casts(), casts(), 0} {
		if got := landed(i); got != want {
			t.Errorf("target %d: %d Chain Lightning outcomes for %d casts, want %d (3 targets in all)", i+1, got, casts(), want)
		}
	}
	t.Logf("%d casts; outcomes per target %d, %d, %d, %d; damage of targets 2 and 3 against the first: %.3f, %.3f",
		casts(), landed(0), landed(1), landed(2), landed(3), m[1].TotalDamage/m[0].TotalDamage, m[2].TotalDamage/m[0].TotalDamage)
	for i, want := range []float64{0.7, 0.49} {
		if got := m[i+1].TotalDamage / m[0].TotalDamage; math.Abs(got-want) > 0.02 {
			t.Errorf("jump %d deals %.3f of the first hit's damage, want %.2f", i+1, got, want)
		}
	}

	var firstHits []float64
	sim.Log = func(message string, vals ...interface{}) {
		if len(vals) > 5 && strings.Contains(message, "BaseDamage") && vals[1] == chainLightningID {
			if base := vals[5].(float64); base > 600 {
				firstHits = append(firstHits, base)
			}
		}
	}
	for i := 0; i < 3000; i++ {
		sim.CurrentTime += 21 * time.Second
		hit(hunter.AutoAttacks.RangedAuto())
	}
	sim.Log = nil
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, base := range firstHits {
		lo, hi = math.Min(lo, base), math.Max(hi, base)
	}
	t.Logf("%d first hits, base damage %.1f to %.1f", len(firstHits), lo, hi)
	if len(firstHits) < 100 || lo < 749 || hi > 853 || hi <= 838 || lo > 775 {
		t.Errorf("%d first hits with base damage %.1f to %.1f, want at least 100 inside 749 to 853 (733 + 4 x 3.9 at level 80, die 105), above the rank's 838 and into the lowest quarter", len(firstHits), lo, hi)
	}

	gear := core.GetGearSet("../../ui/hunter/gear_sets", "p1_mm").GearSet
	gear.Items[proto.ItemSlot_ItemSlotRanged].Id = 150090
	result := core.RunRaidSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Hunter",
				Race:          proto.Race_RaceOrc,
				Class:         proto.Class_ClassHunter,
				Equipment:     gear,
				Consumes:      FullConsumes,
				TalentsString: MMTalents,
				Glyphs:        MMGlyphs,
				Spec:          PlayerOptionsBasic,
				Buffs:         core.FullIndividualBuffs,
				Rotation:      core.GetAplRotation("../../ui/hunter/apls", "mm").Rotation,
			},
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			core.FullDebuffs),
		Encounter:  &proto.Encounter{Duration: 300, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1},
	})
	if result.ErrorResult != "" {
		t.Fatal(result.ErrorResult)
	}
	fightCasts := int32(0)
	for _, action := range result.RaidMetrics.Parties[0].Players[0].Actions {
		if action.Id.GetSpellId() == 25442 {
			for _, tgt := range action.Targets {
				fightCasts += tgt.Casts
			}
		}
	}
	t.Logf("%d Chain Lightning procs in a 300 s fight", fightCasts)
	if fightCasts < 5 || fightCasts > 15 {
		t.Errorf("%d Chain Lightning procs in a 300 s fight, want 5 to 15 (20 s cooldown)", fightCasts)
	}
}
