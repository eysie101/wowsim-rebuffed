package mage

import (
	"math"
	"testing"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/core/stats"
)

func whitemaneArcaneSim() (*core.Simulation, *Mage) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Mage",
				Race:          proto.Race_RaceTroll,
				Class:         proto.Class_ClassMage,
				Equipment:     core.GetGearSet("../../ui/mage/gear_sets", "p3_arcane_alliance").GearSet,
				Consumes:      FullArcaneConsumes,
				TalentsString: ArcaneTalents,
				Glyphs:        ArcaneGlyphs,
				Spec:          PlayerOptionsArcane,
				Buffs:         core.FullIndividualBuffs,
				Rotation:      core.GetAplRotation("../../ui/mage/apls", "arcane").Rotation,
			},
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			core.FullDebuffs),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1},
	})
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*Mage)
}

func TestWhitemaneArcanePotencyArcaneMissiles(t *testing.T) {
	sim, mage := whitemaneArcaneSim()
	target := sim.Encounter.TargetUnits[0]
	potency := mage.ArcanePotencyAura
	bolt := mage.ArcaneMissilesTickSpell
	mage.ArcaneMissiles.BonusHitRating += 100 * core.SpellHitRatingPerHitChance

	potency.Activate(sim)
	if !mage.ArcaneMissiles.Cast(sim, target) {
		t.Fatal("Arcane Missiles not cast")
	}
	if !potency.IsActive() {
		t.Fatal("the Arcane Missiles cast consumed Arcane Potency before its first bolt")
	}
	dot := mage.ArcaneMissiles.Dot(target)

	bolt.BonusHitRating -= 200 * core.SpellHitRatingPerHitChance
	dot.TickOnce(sim)
	if !potency.IsActive() {
		t.Fatal("a bolt that missed consumed Arcane Potency")
	}

	bolt.BonusHitRating += 400 * core.SpellHitRatingPerHitChance
	critWithPotency := bolt.SpellCritChance(target)
	dot.TickOnce(sim)
	if potency.IsActive() {
		t.Fatal("the first bolt that hit left Arcane Potency up")
	}
	if got := critWithPotency - bolt.SpellCritChance(target); math.Abs(got-0.3) > 1e-9 {
		t.Fatalf("the consuming bolt had %+.4f crit chance over the next bolt, want +0.30", got)
	}
}

func TestWhitemaneArcanePotencyBlizzard(t *testing.T) {
	sim, mage := whitemaneArcaneSim()
	potency := mage.ArcanePotencyAura

	potency.Activate(sim)
	if !mage.Blizzard.Cast(sim, sim.Encounter.TargetUnits[0]) {
		t.Fatal("Blizzard not cast")
	}
	dot := mage.Blizzard.AOEDot()
	for tick := 1; tick <= 8; tick++ {
		if !potency.IsActive() {
			t.Fatalf("Arcane Potency gone before Blizzard tick %d", tick)
		}
		dot.TickOnce(sim)
	}
	dot.Cancel(sim)
	if potency.IsActive() {
		t.Fatal("Arcane Potency outlived the Blizzard channel")
	}
}

func TestWhitemaneMirrorImageHit(t *testing.T) {
	sim, mage := whitemaneArcaneSim()
	if mage.Talents.Precision != 3 {
		t.Fatalf("Precision %d, want 3", mage.Talents.Precision)
	}
	mage.mirrorImage.Enable(sim, mage.mirrorImage)
	owner := mage.GetStat(stats.SpellHit)
	want := owner - 3*core.SpellHitRatingPerHitChance
	if got := mage.mirrorImage.GetStat(stats.SpellHit); math.Abs(got-want) > 1e-6 {
		t.Fatalf("images inherit %.2f spell hit rating, want %.2f (the mage's %.2f without Precision's 3%%)", got, want, owner)
	}
}

func TestWhitemaneJudgementOfWisdomMageSpells(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Mage",
				Race:          proto.Race_RaceTroll,
				Class:         proto.Class_ClassMage,
				Equipment:     core.GetGearSet("../../ui/mage/gear_sets", "p3_fire_alliance").GearSet,
				Consumes:      FullFireConsumes,
				TalentsString: FireTalents,
				Glyphs:        FireGlyphs,
				Spec:          PlayerOptionsFire,
				Buffs:         core.FullIndividualBuffs,
				Rotation:      core.GetAplRotation("../../ui/mage/apls", "fire").Rotation,
			},
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			core.FullDebuffs),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1},
	})
	sim.Reset()
	mage := sim.Raid.Parties[0].Players[0].(*Mage)
	for _, c := range []struct {
		name  string
		spell *core.Spell
		procs bool
	}{
		{"Arcane Missiles bolt", mage.ArcaneMissilesTickSpell, false},
		{"Blizzard tick", mage.GetSpell(core.ActionID{SpellID: 42938}), false},
		{"Living Bomb explosion", mage.GetSpell(core.ActionID{SpellID: 55362}), false},
		{"Arcane Missiles cast", mage.ArcaneMissiles, true},
		{"Living Bomb cast", mage.LivingBomb, true},
		{"Fireball", mage.Fireball, true},
	} {
		if c.spell == nil {
			t.Fatalf("%s not registered", c.name)
		}
		if got := core.JudgementOfWisdomManaEvents(sim, c.spell, 200); (got > 0) != c.procs {
			t.Errorf("%s: %d Judgement of Wisdom procs from 200 hits, want procs %v", c.name, got, c.procs)
		}
	}
}
