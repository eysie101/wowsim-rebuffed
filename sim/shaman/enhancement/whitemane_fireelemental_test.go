package enhancement

import (
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/shaman"
)

func whitemaneSummonedFireElemental() (*core.Simulation, *shaman.FireElemental) {
	sim := core.NewSim(whitemaneRequest(1))
	sim.Reset()
	fe := sim.Raid.Parties[0].Players[0].(*EnhancementShaman).FireElemental
	fe.Enable(sim, fe)
	return sim, fe
}

func whitemaneFireElementalAt(t *testing.T, at time.Duration, check func(*core.Simulation, *shaman.FireElemental)) {
	sim, fe := whitemaneSummonedFireElemental()
	done := false
	sim.AddPendingAction(&core.PendingAction{
		NextActionAt: at,
		OnAction: func(sim *core.Simulation) {
			check(sim, fe)
			done = true
		},
	})
	for !done {
		if sim.Step() {
			t.Fatalf("the fight ended before %s", at)
		}
	}
}

func TestWhitemaneFireElementalSwingInterval(t *testing.T) {
	sim, fe := whitemaneSummonedFireElemental()
	melee := fe.AutoAttacks.MHAuto()
	target := fe.CurrentTarget.UnitIndex
	var swings []time.Duration
	for sim.CurrentTime < 8*time.Second {
		casts := melee.SpellMetrics[target].Casts
		if sim.Step() {
			t.Fatal("the fight ended within 8 s")
		}
		if melee.SpellMetrics[target].Casts > casts && sim.CurrentTime < 8*time.Second {
			swings = append(swings, sim.CurrentTime)
		}
	}
	t.Logf("swings before the first Fire Nova of an elemental summoned at 0s: %v", swings)
	if len(swings) < 3 || swings[0] != 0 {
		t.Fatalf("swings %v in the 8 s after a summon at 0s, want at least 3 starting at 0s", swings)
	}
	for i := 1; i < len(swings); i++ {
		if gap := swings[i] - swings[i-1]; gap != 2100*time.Millisecond {
			t.Errorf("swing %d at %s comes %s after the previous one, want 2.1s", i+1, swings[i], gap)
		}
	}
}

func TestWhitemaneFireElementalSpellsTriggerNoGCD(t *testing.T) {
	whitemaneFireElementalAt(t, 3*time.Second, func(sim *core.Simulation, fe *shaman.FireElemental) {
		for _, spell := range []*core.Spell{fe.FireBlast, fe.FireNova} {
			if spell.DefaultCast.GCD != 0 {
				t.Errorf("%s triggers a %s GCD, want none", spell.ActionID, spell.DefaultCast.GCD)
			}
		}
		fe.WaitUntil(sim, sim.CurrentTime)
		if !fe.TryCast(sim, fe.CurrentTarget, fe.FireBlast) {
			t.Fatalf("Fire Blast not cast at %s with the elemental idle", sim.CurrentTime)
		}
		if !fe.GCD.IsReady(sim) {
			t.Errorf("Fire Blast at %s keeps the elemental's GCD running until %s, want no GCD", sim.CurrentTime, fe.GCD.ReadyAt())
		}
		if !fe.TryCast(sim, fe.CurrentTarget, fe.FireNova) {
			t.Fatalf("Fire Nova not castable at %s right after the elemental's Fire Blast", sim.CurrentTime)
		}
		if got, want := fe.GCD.ReadyAt(), sim.CurrentTime+2*time.Second; got != want {
			t.Errorf("Fire Nova at %s keeps the elemental busy until %s, want %s (the end of its 2 s cast)", sim.CurrentTime, got, want)
		}
	})
}

func TestWhitemaneFireElementalInstantBlastKeepsTheSwingTimer(t *testing.T) {
	whitemaneFireElementalAt(t, 3*time.Second, func(sim *core.Simulation, fe *shaman.FireElemental) {
		swingAt := fe.AutoAttacks.NextAttackAt()
		if swingAt <= sim.CurrentTime {
			t.Fatalf("no swing pending at %s, next at %s", sim.CurrentTime, swingAt)
		}
		fe.WaitUntil(sim, sim.CurrentTime)
		if !fe.TryCast(sim, fe.CurrentTarget, fe.FireBlast) {
			t.Fatalf("Fire Blast not cast at %s with the elemental idle", sim.CurrentTime)
		}
		if got := fe.AutoAttacks.NextAttackAt(); got != swingAt {
			t.Errorf("Fire Blast at %s moves the elemental's next swing from %s to %s, want it unchanged", sim.CurrentTime, swingAt, got)
		}
		fe.WaitUntil(sim, sim.CurrentTime)
		if !fe.TryCast(sim, fe.CurrentTarget, fe.FireNova) {
			t.Fatalf("Fire Nova not cast at %s with the elemental idle", sim.CurrentTime)
		}
		if got, want := fe.AutoAttacks.NextAttackAt(), sim.CurrentTime+2*time.Second+fe.AutoAttacks.MainhandSwingSpeed(); got != want {
			t.Errorf("Fire Nova at %s puts the elemental's next swing at %s, want %s (the end of its 2 s cast plus one swing)", sim.CurrentTime, got, want)
		}
	})
}

func TestWhitemaneFireElementalHitScale(t *testing.T) {
	sim, fe := whitemaneSummonedFireElemental()
	target := fe.CurrentTarget
	if n := len(target.DynamicDamageTakenModifiers); n != 0 {
		t.Fatalf("the target has %d post-outcome damage modifiers, which the formula below leaves out", n)
	}
	attackTable := fe.AttackTables[target.UnitIndex]
	const baseDamage = 1000.0
	for _, c := range []struct {
		name   string
		spell  *core.Spell
		factor float64
	}{
		{"melee", fe.AutoAttacks.MHAuto(), 0.929},
		{"Fire Blast", fe.FireBlast, 1.023},
		{"Fire Nova", fe.FireNova, 1.034},
		{"Fire Shield", fe.GetSpell(core.ActionID{SpellID: 13376}), 0.977},
	} {
		if c.spell == nil {
			t.Fatalf("%s: no such spell on the elemental", c.name)
		}
		hit := c.spell.CalcDamage(sim, target, baseDamage, c.spell.OutcomeAlwaysHit)
		formula := baseDamage *
			fe.PseudoStats.DamageDealtMultiplier *
			fe.PseudoStats.SchoolDamageDealtMultiplier[c.spell.SchoolIndex] *
			attackTable.DamageDealtMultiplier *
			c.spell.TargetDamageMultiplier(attackTable, false) *
			hit.ResistanceMultiplier
		got := hit.Damage / formula
		t.Logf("%s: hit %.3f, formula without the factor %.3f, ratio %.4f", c.name, hit.Damage, formula, got)
		if math.Abs(got-c.factor) > 1e-9 {
			t.Errorf("%s: a landed hit deals %.4f x the formula, want %.3f", c.name, got, c.factor)
		}
	}
}

func TestWhitemaneFireElementalSwingsInAFight(t *testing.T) {
	const tolerance = 0.011
	lives, plain, afterNova, badPlain, badAfterNova := 0, 0, 0, 0, 0
	var examples []string
	for seed := int64(1); seed <= 3; seed++ {
		rsr := whitemaneRequest(1)
		rsr.SimOptions.RandomSeed = seed
		rsr.SimOptions.Debug = true
		result := core.RunRaidSim(rsr)
		if result.ErrorResult != "" {
			t.Fatal(result.ErrorResult)
		}
		lastSwing, novaAt := -1.0, -1.0
		for _, line := range strings.Split(result.Logs, "\n") {
			if !strings.Contains(line, "] [Enha (#1) - Greater Fire Elemental] ") {
				continue
			}
			at, err := strconv.ParseFloat(strings.Trim(strings.SplitN(line, " ", 2)[0], "[]"), 64)
			if err != nil {
				t.Fatal(err)
			}
			switch {
			case strings.Contains(line, "Pet summoned"):
				lives++
				lastSwing, novaAt = -1, -1
			case strings.Contains(line, "Casting {SpellID: 12470}"):
				novaAt = at
			case strings.Contains(line, "Casting {OtherID: 3, Tag: 1}"):
				switch {
				case novaAt >= 0:
					afterNova++
					if math.Abs(at-novaAt-4.1) > tolerance {
						badAfterNova++
						examples = append(examples, "seed "+strconv.FormatInt(seed, 10)+": swing "+strconv.FormatFloat(at-novaAt, 'f', 2, 64)+" s after a Fire Nova cast began")
					}
				case lastSwing >= 0:
					plain++
					if math.Abs(at-lastSwing-2.1) > tolerance {
						badPlain++
						examples = append(examples, "seed "+strconv.FormatInt(seed, 10)+": swing "+strconv.FormatFloat(at-lastSwing, 'f', 2, 64)+" s after the previous one")
					}
				}
				lastSwing, novaAt = at, -1
			}
		}
	}
	t.Logf("%d lives: %d swings after a swing, %d swings after a Fire Nova", lives, plain, afterNova)
	if lives < 3 || plain < 60 || afterNova < 30 {
		t.Fatalf("parser not calibrated: %d lives, %d swings after a swing, %d after a Fire Nova in 3 fights", lives, plain, afterNova)
	}
	if badPlain > 0 || badAfterNova > 0 {
		t.Errorf("%d of %d swings not 2.10 s after the previous swing, %d of %d not 4.10 s after a Fire Nova began (2 s cast, then one swing); first: %s",
			badPlain, plain, badAfterNova, afterNova, strings.Join(examples[:min(3, len(examples))], "; "))
	}
}
