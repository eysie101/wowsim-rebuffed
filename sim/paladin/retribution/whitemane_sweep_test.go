package retribution

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func whitemaneRetSim(fanaticism int, mainHand int32) (*core.Simulation, *RetributionPaladin) {
	trees := strings.Split(StandardTalents, "-")
	trees[2] = trees[2][:19] + strconv.Itoa(fanaticism) + trees[2][20:]
	equipment := core.GetGearSet("../../../ui/retribution_paladin/gear_sets", "p1").GearSet
	if mainHand != 0 {
		equipment.Items[proto.ItemSlot_ItemSlotMainHand].Id = mainHand
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Ret",
				Race:          proto.Race_RaceBloodElf,
				Class:         proto.Class_ClassPaladin,
				Equipment:     equipment,
				TalentsString: strings.Join(trees, "-"),
				Glyphs:        StandardGlyphs,
				Consumes:      FullConsumes,
				Spec:          DefaultOptions,
				Buffs:         core.FullIndividualBuffs,
				Rotation:      core.GetAplRotation("../../../ui/retribution_paladin/apls", "default").Rotation,
			},
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			core.FullDebuffs),
		Encounter:  &proto.Encounter{Duration: 300, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1, IsTest: true},
	})
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*RetributionPaladin)
}

func TestWhitemaneFanaticismDivineStormReset(t *testing.T) {
	const n = 3000
	for _, rank := range []int{0, 1, 3} {
		sim, ret := whitemaneRetSim(rank, 130031)
		if ret.Talents.Fanaticism != int32(rank) || !ret.Talents.DivineStorm {
			t.Fatalf("setup: Fanaticism %d, Divine Storm %v; want rank %d and Divine Storm", ret.Talents.Fanaticism, ret.Talents.DivineStorm, rank)
		}
		target := sim.Encounter.TargetUnits[0]
		resets := func(spell *core.Spell, outcome core.HitOutcome) int {
			count := 0
			for i := 0; i < n; i++ {
				ret.DivineStorm.CD.Set(sim.CurrentTime + 10*time.Second)
				result := spell.NewResult(target)
				result.Outcome = outcome
				result.Damage = 1000
				spell.DealDamage(sim, result)
				if ret.DivineStorm.CD.IsReady(sim) {
					count++
				}
			}
			return count
		}
		mhAuto := ret.AutoAttacks.MHAuto()
		armata := ret.GetSpell(core.ActionID{SpellID: 36735})
		if armata == nil {
			t.Fatal("Armata Strigoi's extra attack is not registered")
		}
		want := 0.2
		if rank == 0 {
			want = 0
		}
		for _, c := range []struct {
			name    string
			spell   *core.Spell
			outcome core.HitOutcome
			want    float64
		}{
			{"main-hand swing hit", mhAuto, core.OutcomeHit, want},
			{"main-hand swing crit", mhAuto, core.OutcomeCrit, want},
			{"main-hand swing glance", mhAuto, core.OutcomeGlance, want},
			{"main-hand swing miss", mhAuto, core.OutcomeMiss, 0},
			{"main-hand swing dodge", mhAuto, core.OutcomeDodge, 0},
			{"Crusader Strike hit", ret.CrusaderStrike, core.OutcomeHit, 0},
			{"Divine Storm hit", ret.DivineStorm, core.OutcomeHit, 0},
			{"Armata Strigoi extra attack hit", armata, core.OutcomeHit, 0},
		} {
			share := float64(resets(c.spell, c.outcome)) / n
			if (c.want == 0 && share != 0) || (c.want > 0 && (share < c.want-0.03 || share > c.want+0.03)) {
				t.Errorf("Fanaticism %d/3, %s: %.3f of %d reset Divine Storm's cooldown, want %.2f", rank, c.name, share, n, c.want)
			} else if c.want > 0 {
				t.Logf("Fanaticism %d/3, %s: %.3f of %d reset Divine Storm's cooldown", rank, c.name, share, n)
			}
		}
	}

	sim, ret := whitemaneRetSim(3, 0)
	target := sim.Encounter.TargetUnits[0]
	var casts []time.Duration
	sim.PrePull()
	for !sim.Step() {
		for int(ret.DivineStorm.SpellMetrics[target.UnitIndex].Casts) > len(casts) {
			casts = append(casts, sim.CurrentTime)
		}
	}
	early := 0
	for i := 1; i < len(casts); i++ {
		if casts[i]-casts[i-1] < 10*time.Second {
			early++
		}
	}
	if len(casts) < 20 || early == 0 {
		t.Errorf("300 s fight: %d Divine Storms, %d of them less than 10 s after the previous one; want at least 20 and some early ones", len(casts), early)
	} else {
		t.Logf("300 s fight: %d Divine Storms, %d of them less than 10 s after the previous one", len(casts), early)
	}
}
