package dps

import (
	"testing"
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/deathknight"
)

func whitemaneDk(talents string, items ...*proto.ItemSpec) *proto.Player {
	return &proto.Player{
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassDeathknight,
		Equipment:     &proto.EquipmentSpec{Items: items},
		TalentsString: talents,
		Consumes:      &proto.Consumes{},
		Buffs:         &proto.IndividualBuffs{},
		Spec:          PlayerOptionsFrost,
	}
}

func whitemaneSim(debuffs *proto.Debuffs, players ...*proto.Player) *core.Simulation {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: &proto.Raid{
			Parties: []*proto.Party{{Players: players, Buffs: &proto.PartyBuffs{}}},
			Buffs:   &proto.RaidBuffs{},
			Debuffs: debuffs,
		},
		Encounter: &proto.Encounter{
			Duration: 60,
			Targets:  []*proto.Target{{Level: 80, MobType: proto.MobType_MobTypeUndead}},
		},
		SimOptions: &proto.SimOptions{RandomSeed: 1, IsTest: true},
	})
	sim.Reset()
	return sim
}

func whitemaneDkOf(sim *core.Simulation, index int) *deathknight.Deathknight {
	return sim.Raid.Parties[0].Players[index].(*DpsDeathknight).Deathknight
}

func TestWhitemaneBloodBoilAnyDeathKnight(t *testing.T) {
	ep := &proto.Debuffs{EbonPlaguebringer: true}
	bloodBoil := func(sim *core.Simulation, spell *core.Spell) float64 {
		target := sim.Encounter.TargetUnits[0]
		spell.Cast(sim, target)
		return spell.SpellMetrics[target.UnitIndex].TotalDamage
	}

	clean := whitemaneSim(ep, whitemaneDk(""), whitemaneDk(""))
	base := bloodBoil(clean, whitemaneDkOf(clean, 1).BloodBoil)

	other := whitemaneSim(ep, whitemaneDk(""), whitemaneDk(""))
	whitemaneDkOf(other, 0).FrostFeverSpell.Dot(other.Encounter.TargetUnits[0]).Apply(other)

	solo := whitemaneSim(ep, whitemaneDk(""))

	drwClean := whitemaneSim(&proto.Debuffs{}, whitemaneDk(BloodTalents))
	drwOwner := whitemaneSim(&proto.Debuffs{}, whitemaneDk(BloodTalents))
	whitemaneDkOf(drwOwner, 0).FrostFeverSpell.Dot(drwOwner.Encounter.TargetUnits[0]).Apply(drwOwner)
	runeWeaponBloodBoil := func(sim *core.Simulation) float64 {
		runeWeapon := whitemaneDkOf(sim, 0).RuneWeapon
		runeWeapon.EnableWithTimeout(sim, runeWeapon, 12*time.Second)
		return bloodBoil(sim, runeWeapon.BloodBoil)
	}
	drwBase := runeWeaponBloodBoil(drwClean)

	if base == 0 || drwBase == 0 {
		t.Fatalf("Blood Boil did not land on the undiseased target (dk %.1f, rune weapon %.1f)", base, drwBase)
	}
	for name, ratio := range map[string]float64{
		"another Death Knight's Frost Fever":        bloodBoil(other, whitemaneDkOf(other, 1).BloodBoil) / base,
		"an external Ebon Plaguebringer, single DK": bloodBoil(solo, whitemaneDkOf(solo, 0).BloodBoil) / base,
		"rune weapon, the owner's Frost Fever":      runeWeaponBloodBoil(drwOwner) / drwBase,
	} {
		if !core.WithinToleranceFloat64(1.5, ratio, 1e-9) {
			t.Errorf("%s: Blood Boil deals %.4f times the undiseased hit, want 1.5", name, ratio)
		}
	}
}

func TestWhitemaneThreatOfThassarianFullFlatBonus(t *testing.T) {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotOffHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 47528}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: 47528}
	sim := whitemaneSim(&proto.Debuffs{}, whitemaneDk(FrostTalents, items...))
	dk := whitemaneDkOf(sim, 0)
	target := sim.Encounter.TargetUnits[0]

	offHand := *dk.AutoAttacks.OH()
	offHand.BaseDamageMin, offHand.BaseDamageMax, offHand.NormalizedSwingSpeed = 0, 0, 0
	dk.AutoAttacks.SetOH(offHand)

	for _, strike := range []struct {
		name  string
		spell *core.Spell
		flat  float64
	}{
		{"Death Strike", dk.DeathStrikeOhHit, 297},
		{"Obliterate", dk.ObliterateOhHit, 584},
		{"Plague Strike", dk.PlagueStrikeOhHit, 378},
		{"Blood Strike", dk.BloodStrikeOhHit, 764},
		{"Frost Strike", dk.FrostStrikeOhHit, 250},
	} {
		strike.spell.BonusCritRating = -100 * core.CritRatingPerCritChance
		perPoint := strike.spell.CalcDamage(sim, target, 1, strike.spell.OutcomeAlwaysHit).Damage
		strike.spell.Cast(sim, target)
		if got := strike.spell.SpellMetrics[target.UnitIndex].TotalDamage / perPoint; !core.WithinToleranceFloat64(strike.flat, got, 1e-6) {
			t.Errorf("off-hand %s: flat bonus %.2f, want the main hand's %.0f", strike.name, got, strike.flat)
		}
	}
}

func TestWhitemaneArmyGhoulClawCadence(t *testing.T) {
	player := whitemaneDk("")
	player.Rotation = core.APLRotationFromJsonString(`{"type":"TypeAPL","priorityList":[{"action":{"castSpell":{"spellId":{"spellId":42650}}}}]}`)
	sim := whitemaneSim(&proto.Debuffs{}, player)
	dk := whitemaneDkOf(sim, 0)
	target := sim.Encounter.TargetUnits[0]

	summonedAt := make([]time.Duration, len(dk.ArmyGhoul))
	claws := make([][]time.Duration, len(dk.ArmyGhoul))
	sim.PrePull()
	for !sim.Step() {
		for i, ghoul := range dk.ArmyGhoul {
			if ghoul.IsEnabled() && summonedAt[i] == 0 {
				summonedAt[i] = sim.CurrentTime
			}
			for int(ghoul.Claw.SpellMetrics[target.UnitIndex].Casts) > len(claws[i]) {
				claws[i] = append(claws[i], sim.CurrentTime)
			}
		}
	}

	for i, times := range claws {
		if len(times) < 7 {
			t.Errorf("army ghoul %d: %d Claws in its 40 s, want at least 7", i, len(times))
		}
		previous := summonedAt[i]
		for _, at := range times {
			if gap := at - previous; gap < 4*time.Second || gap > 5300*time.Millisecond {
				t.Errorf("army ghoul %d: Claw at %s, %s after the summon or the previous Claw, want 4 to 5.3 s", i, at, gap)
			}
			previous = at
		}
	}
}
