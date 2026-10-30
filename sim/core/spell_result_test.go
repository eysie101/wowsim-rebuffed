package core

import (
	"fmt"
	"strings"
	"testing"

	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/core/stats"
)

func TestCalcDamageEqualWithLogging(t *testing.T) {
	attacker := &Unit{Type: PlayerUnit, Level: 80, PseudoStats: stats.NewPseudoStats()}
	target := &Unit{Type: EnemyUnit, Level: 83, PseudoStats: stats.NewPseudoStats()}
	attacker.PseudoStats.DamageDealtMultiplier = 1.2
	target.PseudoStats.DamageTakenMultiplier = 1.3
	target.PseudoStats.BonusPhysicalDamageTaken = 1000
	target.stats[stats.Armor] = 10643
	target.DynamicDamageTakenModifiers = []DynamicDamageTakenModifier{func(_ *Simulation, _ *Spell, result *SpellResult) { result.Damage -= 100 }}
	attacker.AttackTables = []*AttackTable{NewAttackTable(attacker, target)}
	spell := &Spell{Unit: attacker, SpellSchool: SpellSchoolPhysical, Flags: SpellFlagIncludeTargetBonusDamage, DamageMultiplier: 1, DamageMultiplierAdditive: 1}
	crit := func(_ *Simulation, result *SpellResult, _ *AttackTable) {
		result.Outcome = OutcomeCrit
		result.Damage *= 2
	}

	sim := NewSim(&proto.RaidSimRequest{SimOptions: &proto.SimOptions{}, Encounter: &proto.Encounter{}, Raid: &proto.Raid{}})
	off := *spell.CalcDamage(sim, target, 2000, crit)
	var debug string
	sim.Log = func(format string, vals ...interface{}) { debug = fmt.Sprintf(format, vals...) }
	on := *spell.CalcDamage(sim, target, 2000, crit)

	if on.Damage != off.Damage || on.PreOutcomeDamage != off.PreOutcomeDamage {
		t.Errorf("with logging: damage %f, pre-outcome damage %f; without: %f, %f; want equal", on.Damage, on.PreOutcomeDamage, off.Damage, off.PreOutcomeDamage)
	}
	want := fmt.Sprintf("AfterAttackerMods:%0.01f, AfterTargetMods:%0.01f, AfterResistances:%0.01f, AfterOutcome:%0.01f, AfterPostOutcome:%0.01f",
		2000*1.2, (2000*1.2+1000)*1.3, off.PreOutcomeDamage, off.PreOutcomeDamage*2, off.Damage)
	if !strings.HasSuffix(debug, want) {
		t.Errorf("debug line %q, want it to end with %q", debug, want)
	}
}
