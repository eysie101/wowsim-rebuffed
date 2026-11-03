package hunter

import (
	"math"
	"strconv"
	"testing"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func whitemaneSweepHunterSim(talents string) (*core.Simulation, *Hunter) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Hunter",
				Race:          proto.Race_RaceOrc,
				Class:         proto.Class_ClassHunter,
				Equipment:     core.GetGearSet("../../ui/hunter/gear_sets", "p1_sv").GearSet,
				Consumes:      FullConsumes,
				TalentsString: talents,
				Glyphs:        BMGlyphs,
				Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{
					Ammo:       proto.Hunter_Options_SaroniteRazorheads,
					PetType:    proto.Hunter_Options_Cat,
					PetTalents: FerocityTalents,
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
	return sim, sim.Raid.Parties[0].Players[0].(*Hunter)
}

func TestWhitemaneAnimalHandlerMana(t *testing.T) {
	for _, c := range []struct {
		animalHandler, invigoration     int32
		crits, minRestores, maxRestores int
	}{{2, 0, 50, 50, 50}, {1, 0, 50, 13, 37}, {0, 0, 50, 0, 0}, {0, 2, 40, 40, 40}, {2, 2, 40, 80, 80}} {
		talents := BMTalents[:14] + strconv.Itoa(int(c.animalHandler)) + BMTalents[15:19] + strconv.Itoa(int(c.invigoration)) + BMTalents[20:]
		sim, hunter := whitemaneSweepHunterSim(talents)
		if hunter.Talents.AnimalHandler != c.animalHandler || hunter.Talents.Invigoration != c.invigoration {
			t.Fatalf("talents give Animal Handler %d and Invigoration %d, want %d and %d", hunter.Talents.AnimalHandler, hunter.Talents.Invigoration, c.animalHandler, c.invigoration)
		}
		pet := hunter.pet
		target := sim.Encounter.TargetUnits[0]
		hunter.SpendMana(sim, hunter.CurrentMana(), hunter.NewManaMetrics(core.ActionID{SpellID: 1}))
		onePercent := 0.01 * hunter.MaxMana()

		pet.OnSpellHitDealt(sim, pet.AutoAttacks.MHAuto(), &core.SpellResult{Target: target, Outcome: core.OutcomeCrit})
		pet.OnSpellHitDealt(sim, pet.focusDump, &core.SpellResult{Target: target, Outcome: core.OutcomeHit})
		if got := hunter.CurrentMana(); got != 0 {
			t.Errorf("Animal Handler %d, Invigoration %d: a white crit and a special hit restore %.1f mana, want 0", c.animalHandler, c.invigoration, got)
		}

		for i := 0; i < c.crits; i++ {
			pet.OnSpellHitDealt(sim, pet.focusDump, &core.SpellResult{Target: target, Outcome: core.OutcomeCrit})
		}
		restores := hunter.CurrentMana() / onePercent
		if math.Abs(restores-math.Round(restores)) > 1e-6 || int(math.Round(restores)) < c.minRestores || int(math.Round(restores)) > c.maxRestores {
			t.Errorf("Animal Handler %d, Invigoration %d: %d special crits restore %.3f x 1%% of maximum mana, want %d to %d whole restores", c.animalHandler, c.invigoration, c.crits, restores, c.minRestores, c.maxRestores)
		}
	}
}
