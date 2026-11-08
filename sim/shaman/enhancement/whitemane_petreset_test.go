package enhancement

import (
	"math"
	"testing"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/stats"
	"github.com/wowsims/wotlk/sim/shaman"
)

func TestWhitemaneSpiritWolvesKeepOwnerShareAfterReset(t *testing.T) {
	rsr := whitemaneRequest(1)
	rsr.Encounter.Duration = 196
	sim := core.NewSim(rsr)
	enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
	wolves := []*shaman.SpiritWolf{enh.SpiritWolves.SpiritWolf1, enh.SpiritWolves.SpiritWolf2}
	base := make([]stats.Stats, len(wolves))
	share := make([]float64, len(wolves))
	for iteration := 1; iteration <= 4; iteration++ {
		sim.Reset()
		for i, wolf := range wolves {
			if iteration == 1 {
				base[i] = wolf.GetStats()
			} else if got := wolf.GetStats(); !got.Equals(base[i]) {
				t.Errorf("iteration %d: %s starts with %.1f attack power, %.1f melee hit and %.1f expertise rating, want %.1f, %.1f and %.1f as in iteration 1", iteration, wolf.Label, got[stats.AttackPower], got[stats.MeleeHit], got[stats.Expertise], base[i][stats.AttackPower], base[i][stats.MeleeHit], base[i][stats.Expertise])
			}
		}
		sim.PrePull()
		for !sim.Step() {
		}
		ownerAP := enh.GetStat(stats.AttackPower)
		for i, wolf := range wolves {
			if !wolf.IsEnabled() {
				t.Fatalf("iteration %d: %s is not summoned when the fight ends; the test needs the second Feral Spirit still out", iteration, wolf.Label)
			}
			got := (wolf.GetStat(stats.AttackPower) - base[i][stats.AttackPower]) / ownerAP
			t.Logf("iteration %d: %s %.1f attack power, %.4f of the shaman's %.1f above its base", iteration, wolf.Label, wolf.GetStat(stats.AttackPower), got, ownerAP)
			if iteration == 1 {
				if got < 0.5 {
					t.Fatalf("iteration 1: %s carries %.3f of the shaman's attack power, want at least the glyphed 0.61", wolf.Label, got)
				}
				share[i] = got
			} else if math.Abs(got-share[i]) > 0.01*share[i] {
				t.Errorf("iteration %d: %s carries %.3f of the shaman's attack power when the fight ends, want %.3f as in iteration 1", iteration, wolf.Label, got, share[i])
			}
		}
		sim.Cleanup()
	}
}
