package dps

import (
	"math"
	"strings"
	"testing"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func armsRequest(seed int64, numTargets int) *proto.RaidSimRequest {
	targets := make([]*proto.Target, numTargets)
	for i := range targets {
		targets[i] = core.NewDefaultTarget()
	}
	return &proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Arms",
				Race:          proto.Race_RaceOrc,
				Class:         proto.Class_ClassWarrior,
				Equipment:     core.GetGearSet("../../../ui/warrior/gear_sets", "p1_arms").GearSet,
				Consumes:      FullConsumes,
				Spec:          PlayerOptionsArms,
				TalentsString: ArmsTalents,
				Glyphs:        ArmsGlyphs,
				Buffs:         core.FullIndividualBuffs,
				Rotation:      core.GetAplRotation("../../../ui/warrior/apls", "arms").Rotation,
			},
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			core.FullDebuffs),
		Encounter: &proto.Encounter{
			Duration:             300,
			ExecuteProportion_20: 0.2,
			ExecuteProportion_25: 0.25,
			ExecuteProportion_35: 0.35,
			Targets:              targets,
		},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: seed, Debug: true},
	}
}

func TestWhitemaneSuddenDeathFromExecute(t *testing.T) {
	consumed, kept, lost := 0, 0, 0
	for seed := int64(1); seed <= 10; seed++ {
		result := core.RunRaidSim(armsRequest(seed, 1))
		if result.ErrorResult != "" {
			t.Fatal(result.ErrorResult)
		}
		active, open, before, refreshed, stamp := false, false, false, false, ""
		closeExecute := func() {
			if open && before {
				consumed++
				if active {
					kept++
				}
			}
			open = false
		}
		for _, line := range strings.Split(result.Logs, "\n") {
			if !strings.Contains(line, "] [Arms (#1)] ") || strings.Contains(line, "[DEBUG]") {
				continue
			}
			if s := strings.SplitN(line, " ", 2)[0]; s != stamp {
				closeExecute()
				stamp = s
			}
			switch {
			case strings.Contains(line, "{SpellID: 47471} Hit for") || strings.Contains(line, "{SpellID: 47471} Crit for"):
				closeExecute()
				open, before, refreshed = true, active, false
			case strings.Contains(line, "Aura gained: {SpellID: 29724}"):
				active = true
			case strings.Contains(line, "Aura refreshed: {SpellID: 29724}"):
				refreshed = refreshed || open
			case strings.Contains(line, "Aura faded: {SpellID: 29724}"):
				active = false
				if open && before && refreshed {
					lost++
				}
			}
		}
		closeExecute()
	}
	if consumed < 50 {
		t.Fatalf("parser not calibrated: only %d landed Executes used a Sudden Death in 10 fights", consumed)
	}
	if share := float64(kept) / float64(consumed); lost > 0 || share < 0.02 || share > 0.25 {
		t.Fatalf("of %d landed Executes that used a Sudden Death, %d left a new one active and %d procced one that the same Execute then removed; with 3/3 Sudden Death about 9%% should leave a new one", consumed, kept, lost)
	}
	t.Logf("%d landed Executes used a Sudden Death, %d of them procced a new one", consumed, kept)
}

func TestWhitemaneSweepingStrikesCopies(t *testing.T) {
	rsr := armsRequest(1, 3)
	rsr.SimOptions.Debug = false
	rsr.SimOptions.IsTest = true
	rsr.Encounter.ExecuteProportion_20, rsr.Encounter.ExecuteProportion_25, rsr.Encounter.ExecuteProportion_35 = 1, 1, 1
	sim := core.NewSim(rsr)
	sim.Reset()
	sim.PrePull()
	for !sim.IsExecutePhase20() && !sim.Step() {
	}
	if !sim.IsExecutePhase20() {
		t.Fatal("the sim never reached the execute phase")
	}
	war := sim.Raid.Parties[0].Players[0].(*DpsWarrior).Warrior
	ss := war.GetAura("Sweeping Strikes")
	ssHit := war.GetSpell(core.ActionID{SpellID: 12723})
	if ss == nil || ssHit == nil || war.Bladestorm == nil {
		t.Fatal("no Sweeping Strikes or Bladestorm on the Arms warrior")
	}
	target, next := sim.Encounter.TargetUnits[0], sim.Encounter.TargetUnits[1]
	copyOf := func(spell *core.Spell) (float64, int32) {
		ss.Deactivate(sim)
		ss.Activate(sim)
		before := ssHit.SpellMetrics[next.UnitIndex].TotalDamage
		ss.OnSpellHitDealt(ss, sim, spell, &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100000, ResistanceMultiplier: 1})
		return ssHit.SpellMetrics[next.UnitIndex].TotalDamage - before, ss.GetStacks()
	}
	if damage, stacks := copyOf(war.MortalStrike); damage < 50000 || stacks != 4 {
		t.Fatalf("Mortal Strike for 100000: copy %.0f, %d charges left; want a copy of the hit and 4 charges", damage, stacks)
	}
	if damage, stacks := copyOf(war.Execute); damage <= 0 || damage > 10000 || stacks != 4 {
		t.Errorf("Execute for 100000 in the execute phase: copy %.0f, %d charges left; want a normalized weapon hit and 4 charges", damage, stacks)
	} else {
		t.Logf("Execute for 100000 in the execute phase: copy %.0f", damage)
	}
	for _, spell := range []*core.Spell{war.Whirlwind, war.Bladestorm} {
		if damage, stacks := copyOf(spell); damage != 0 || stacks != 5 {
			t.Errorf("%s for 100000: copy %.0f, %d charges left; want no copy and 5 charges", spell.ActionID, damage, stacks)
		}
	}
}

func TestWhitemaneSweepingStrikesExecuteCopyArmor(t *testing.T) {
	rsr := armsRequest(1, 3)
	rsr.SimOptions.Debug = false
	rsr.SimOptions.IsTest = true
	rsr.Encounter.ExecuteProportion_20, rsr.Encounter.ExecuteProportion_25, rsr.Encounter.ExecuteProportion_35 = 1, 1, 1
	sim := core.NewSim(rsr)
	sim.Reset()
	sim.PrePull()
	for !sim.IsExecutePhase20() && !sim.Step() {
	}
	if !sim.IsExecutePhase20() {
		t.Fatal("the sim never reached the execute phase")
	}
	war := sim.Raid.Parties[0].Players[0].(*DpsWarrior).Warrior
	ss := war.GetAura("Sweeping Strikes")
	ssHit := war.GetSpell(core.ActionID{SpellID: 12723})
	if ss == nil || ssHit == nil {
		t.Fatal("no Sweeping Strikes on the Arms warrior")
	}
	target, next := sim.Encounter.TargetUnits[0], sim.Encounter.TargetUnits[1]
	const n = 400
	meanCopy := func(firstArmor float64) float64 {
		total := 0.0
		for i := 0; i < n; i++ {
			ss.Deactivate(sim)
			ss.Activate(sim)
			before := ssHit.SpellMetrics[next.UnitIndex].TotalDamage
			ss.OnSpellHitDealt(ss, sim, war.Execute, &core.SpellResult{Target: target, Outcome: core.OutcomeHit, Damage: 100000, ResistanceMultiplier: firstArmor})
			total += ssHit.SpellMetrics[next.UnitIndex].TotalDamage - before
		}
		return total / n
	}
	casts := func() (int32, int32) {
		return ssHit.SpellMetrics[target.UnitIndex].Casts, ssHit.SpellMetrics[next.UnitIndex].Casts
	}
	firstCasts, nextCasts := casts()
	plain := meanCopy(1)
	if c1, c2 := casts(); c1 != firstCasts || c2 != nextCasts {
		t.Errorf("%d copies changed the copy spell's casts by %d on the first target and %d on the second; want 0 and 0", n, c1-firstCasts, c2-nextCasts)
	}
	next.PseudoStats.BonusPhysicalDamageTaken += 1000
	bonus := meanCopy(1) - plain
	next.PseudoStats.BonusPhysicalDamageTaken -= 1000
	if bonus < 300 || bonus > 1100 {
		t.Errorf("1000 more flat physical damage taken on the second target raises the Execute copy by %.0f; want it added as on any weapon hit", bonus)
	}
	if r := meanCopy(0.5) / plain; r < 0.95 || r > 1.05 {
		t.Errorf("Execute copy with the first target's armor multiplier 0.5 is %.3f x the copy at 1; want 1, the copy takes only the second target's armor", r)
	}
	war.PseudoStats.DamageDealtMultiplier *= 2
	doubled := meanCopy(1)
	war.PseudoStats.DamageDealtMultiplier /= 2
	if r := doubled / plain; r < 1.9 || r > 2.1 {
		t.Errorf("Execute copy with twice the warrior's damage multiplier is %.3f x the plain copy; want 2, a normalized weapon hit takes the warrior's damage modifiers", r)
	}
	raw := 0.0
	for i := 0; i < n; i++ {
		raw += war.MHNormalizedWeaponDamage(sim, war.Execute.MeleeAttackPower()) + war.Execute.BonusWeaponDamage()
	}
	raw /= n
	if r := plain / raw; r < 0.4 || r > 0.95 {
		t.Errorf("Execute copy %.0f is %.3f x the raw normalized weapon hit %.0f; want the second target's armor applied", plain, r, raw)
	} else {
		t.Logf("Execute copy %.0f against the raw normalized weapon hit %.0f (%.3f)", plain, raw, r)
	}
}

func TestWhitemaneThunderClapCrit(t *testing.T) {
	rsr := armsRequest(1, 1)
	env, _, _ := core.NewEnvironment(rsr.Raid, rsr.Encounter, false)
	war := env.Raid.Parties[0].Players[0].(*DpsWarrior).Warrior
	if war.Talents.Cruelty != 5 || war.Talents.PoleaxeSpecialization != 5 || war.Talents.Incite != 3 || war.MainHand().WeaponType != proto.WeaponType_WeaponTypeAxe {
		t.Fatalf("setup: Cruelty %d, Poleaxe Specialization %d, Incite %d, main hand %s; want 5, 5, 3 and an axe", war.Talents.Cruelty, war.Talents.PoleaxeSpecialization, war.Talents.Incite, war.MainHand().WeaponType)
	}
	attackTable := war.AttackTables[env.Encounter.TargetUnits[0].UnitIndex]
	if gap := war.ThunderClap.PhysicalCritChance(attackTable) - war.Execute.PhysicalCritChance(attackTable); math.Abs(gap-0.05) > 1e-9 {
		t.Fatalf("Thunder Clap crit chance %+.4f against Execute; want +0.0500 (Incite +15%%, without Cruelty 5%% and Poleaxe Specialization 5%%)", gap)
	}
}
