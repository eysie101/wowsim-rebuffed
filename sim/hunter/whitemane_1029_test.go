package hunter

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func whitemanePetSim(petType proto.Hunter_Options_PetType, petTalents *proto.HunterPetTalents) (*core.Simulation, *HunterPet) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Hunter",
				Race:          proto.Race_RaceOrc,
				Class:         proto.Class_ClassHunter,
				Equipment:     core.GetGearSet("../../ui/hunter/gear_sets", "p1_sv").GearSet,
				Consumes:      FullConsumes,
				TalentsString: BMTalents,
				Glyphs:        BMGlyphs,
				Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{
					Ammo:       proto.Hunter_Options_SaroniteRazorheads,
					PetType:    petType,
					PetTalents: petTalents,
					PetUptime:  1,
				}}},
				Buffs:    core.FullIndividualBuffs,
				Rotation: core.GetAplRotation("../../ui/hunter/apls", "bm").Rotation,
			},
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			core.FullDebuffs),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1},
	})
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*Hunter).pet
}

func TestWhitemanePetFocusAbilityGCD(t *testing.T) {
	for _, petType := range []proto.Hunter_Options_PetType{proto.Hunter_Options_Cat, proto.Hunter_Options_Boar, proto.Hunter_Options_Wasp} {
		for _, special := range []bool{false, true} {
			sim, pet := whitemanePetSim(petType, FerocityTalents)
			spell := pet.focusDump
			if special {
				spell = pet.specialAbility
			}
			if !spell.Cast(sim, sim.Encounter.TargetUnits[0]) {
				t.Fatalf("%s: %s not cast", pet.Name, spell.ActionID)
			}
			if got := pet.GCD.ReadyAt() - sim.CurrentTime; got != 1500*time.Millisecond {
				t.Errorf("%s: %s starts a %s pet GCD, want 1.5s", pet.Name, spell.ActionID, got)
			}
		}
	}
}

func TestWhitemaneFocusedFireKillCommand(t *testing.T) {
	sim, pet := whitemanePetSim(proto.Hunter_Options_Raptor, &proto.HunterPetTalents{WolverineBite: true})
	ids := []int32{ClawSpellID, 53582, 53508}
	base := map[int32]float64{}
	for _, id := range ids {
		spell := pet.GetSpell(core.ActionID{SpellID: id})
		if spell == nil {
			t.Fatalf("pet spell %d not registered", id)
		}
		base[id] = spell.BonusCritRating
	}
	check := func(when string, want float64) {
		for _, id := range ids {
			if got := pet.GetSpell(core.ActionID{SpellID: id}).BonusCritRating - base[id]; math.Abs(got-want) > 1e-6 {
				t.Errorf("spell %d %s: %+.2f%% crit, want %+.2f%%", id, when, got/core.CritRatingPerCritChance, want/core.CritRatingPerCritChance)
			}
		}
	}
	pet.KillCommandAura.Activate(sim)
	pet.KillCommandAura.SetStacks(sim, 3)
	check("while Kill Command is up", 20*core.CritRatingPerCritChance)
	pet.KillCommandAura.SetStacks(sim, 0)
	check("after Kill Command", 0)
}

func TestWhitemaneJudgementOfWisdomHunterSpells(t *testing.T) {
	gear := core.GetGearSet("../../ui/hunter/gear_sets", "p1_mm").GearSet
	gear.Items[16].Id = 50034
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
			core.FullDebuffs),
		Encounter:  &proto.Encounter{Duration: 60, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1},
	})
	sim.Reset()
	hunter := sim.Raid.Parties[0].Players[0].(*Hunter)
	for _, c := range []struct {
		name  string
		spell *core.Spell
		procs bool
	}{
		{"Chimera Shot - Serpent", hunter.GetSpell(core.ActionID{SpellID: 53353}), false},
		{"Wild Quiver", hunter.GetSpell(core.ActionID{SpellID: 53217}), false},
		{"Explosive Trap", hunter.ExplosiveTrap, false},
		{"Zod's Repeating Longbow", hunter.GetSpell(core.ActionID{ItemID: 50034}), false},
		{"Auto Shot", hunter.AutoAttacks.RangedAuto(), true},
		{"Steady Shot", hunter.SteadyShot, true},
		{"Chimera Shot", hunter.ChimeraShot, true},
	} {
		if c.spell == nil {
			t.Fatalf("%s not registered", c.name)
		}
		if got := core.JudgementOfWisdomManaEvents(sim, c.spell, 200); (got > 0) != c.procs {
			t.Errorf("%s: %d Judgement of Wisdom procs from 200 hits, want procs %v", c.name, got, c.procs)
		}
	}
}
