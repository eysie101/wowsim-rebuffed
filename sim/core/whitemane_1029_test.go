package core

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/core/stats"
)

func init() {
	RegisterAgentFactory(
		proto.Player_Warlock{},
		proto.Spec_SpecWarlock,
		newWhitemaneAutocastOwner,
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_Warlock)
		},
	)
}

type whitemaneAutocastOwner struct {
	Character
	pet *whitemaneAutocastPet
}

func (owner *whitemaneAutocastOwner) GetCharacter() *Character { return &owner.Character }
func (owner *whitemaneAutocastOwner) Initialize()              {}
func (owner *whitemaneAutocastOwner) ApplyTalents()            {}
func (owner *whitemaneAutocastOwner) Reset(_ *Simulation)      {}

type whitemaneAutocastPet struct {
	Pet
	strike *Spell
	starts []time.Duration
}

func (pet *whitemaneAutocastPet) GetPet() *Pet        { return &pet.Pet }
func (pet *whitemaneAutocastPet) Reset(_ *Simulation) {}
func (pet *whitemaneAutocastPet) Initialize() {
	pet.strike = pet.RegisterSpell(SpellConfig{
		ActionID:    ActionID{SpellID: 1},
		SpellSchool: SpellSchoolPhysical,
		ProcMask:    ProcMaskEmpty,
		Cast: CastConfig{
			DefaultCast: Cast{GCD: time.Millisecond * 1500},
			IgnoreHaste: true,
		},
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *Simulation, _ *Unit, _ *Spell) {
			pet.starts = append(pet.starts, sim.CurrentTime)
		},
	})
}
func (pet *whitemaneAutocastPet) ExecuteCustomRotation(sim *Simulation) {
	pet.strike.Cast(sim, pet.CurrentTarget)
}

func newWhitemaneAutocastOwner(character *Character, _ *proto.Player) Agent {
	owner := &whitemaneAutocastOwner{Character: *character}
	owner.pet = &whitemaneAutocastPet{
		Pet: NewAutocastPet("Autocast Pet", &owner.Character, stats.Stats{}, func(stats.Stats) stats.Stats { return stats.Stats{} }, true, false),
	}
	owner.pet.EnableAutoAttacks(owner.pet, AutoAttackOptions{
		MainHand:       Weapon{BaseDamageMin: 1, BaseDamageMax: 1, SwingSpeed: 1, CritMultiplier: 2},
		AutoSwingMelee: true,
	})
	owner.AddPet(owner.pet)
	return owner
}

func TestWhitemaneAutocastGrid(t *testing.T) {
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{{
			Name:      "Owner",
			Class:     proto.Class_ClassWarlock,
			Spec:      &proto.Player_Warlock{},
			Consumes:  &proto.Consumes{},
			Buffs:     &proto.IndividualBuffs{},
			Equipment: &proto.EquipmentSpec{},
		}}, Buffs: &proto.PartyBuffs{}}}},
		Encounter: &proto.Encounter{
			Duration: 30,
			Targets:  []*proto.Target{{Name: "target", Level: 83}},
		},
	})
	pet := sim.Raid.Parties[0].Players[0].(*whitemaneAutocastOwner).pet
	sim.Reset()
	sim.PrePull()
	for !sim.Step() {
	}
	sim.Cleanup()

	t.Logf("%d strikes, first starts %v", len(pet.starts), pet.starts[:min(6, len(pet.starts))])
	if len(pet.starts) < 15 {
		t.Fatalf("%d strikes in 30 s, want at least 15", len(pet.starts))
	}
	for i, start := range pet.starts {
		if offset := start % (800 * time.Millisecond); offset != 0 {
			t.Errorf("strike %d starts at %s, %s after the last 0.8 s grid point", i+1, start, offset)
		}
		if i > 0 && start-pet.starts[i-1] < 1500*time.Millisecond {
			t.Errorf("strike %d starts %s after the previous one, inside the 1.5 s GCD", i+1, start-pet.starts[i-1])
		}
	}
}

const whitemaneResistSamples = 200_000

func whitemaneResistSim() *Simulation {
	return NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Encounter:  &proto.Encounter{},
		Raid:       &proto.Raid{},
	})
}

func whitemanePartiallyResistedShare(sim *Simulation, attackTable *AttackTable) float64 {
	spell := &Spell{SpellSchool: SpellSchoolFire}
	partial := 0
	for i := 0; i < whitemaneResistSamples; i++ {
		result := SpellResult{Outcome: OutcomeHit, Damage: 1000}
		result.applyResistances(sim, spell, false, attackTable)
		if result.ResistanceMultiplier < 1 {
			partial++
		}
	}
	return float64(partial) / whitemaneResistSamples
}

func TestWhitemanePartialResistVsBoss(t *testing.T) {
	sim := whitemaneResistSim()
	player := &Unit{Type: PlayerUnit, Level: 80}
	boss := &Unit{Type: EnemyUnit, Level: 83}

	share := whitemanePartiallyResistedShare(sim, NewAttackTable(player, boss))
	t.Logf("partially resisted hits against a +3 boss: %.2f%%", share*100)
	if math.Abs(share-0.45) > 0.005 {
		t.Errorf("partially resisted share against a +3 boss %.4f, want 0.45", share)
	}

	player.stats[stats.SpellPenetration] = 1000
	share = whitemanePartiallyResistedShare(sim, NewAttackTable(player, boss))
	t.Logf("with 1000 spell penetration: %.2f%%", share*100)
	if math.Abs(share-0.45) > 0.005 {
		t.Errorf("partially resisted share with spell penetration %.4f, want 0.45", share)
	}

	for _, defender := range []*Unit{{Type: PlayerUnit, Level: 83}, {Type: PetUnit, Level: 83}} {
		share = whitemanePartiallyResistedShare(sim, NewAttackTable(&Unit{Type: EnemyUnit, Level: 80}, defender))
		t.Logf("against a level 83 unit of type %d: %.2f%%", defender.Type, share*100)
		if share != 0 {
			t.Errorf("unit type %d is partially resisted %.4f of the time, want 0", defender.Type, share)
		}
	}
}

func TestWhitemanePartialResistOldServerControl(t *testing.T) {
	sim := whitemaneResistSim()
	thresholds := (&Unit{}).partialResistRollThresholds(3 * 5.0 / (3*5.0 + 510))
	partial := 0
	for i := 0; i < whitemaneResistSamples; i++ {
		roll := sim.RandomFloat("Partial Resist")
		for _, threshold := range thresholds {
			if roll < threshold.cumulativeChance {
				if threshold.bracket > 0 {
					partial++
				}
				break
			}
		}
	}
	share := float64(partial) / whitemaneResistSamples
	t.Logf("old server model (5 resistance per level, constant 510): %.2f%%", share*100)
	if math.Abs(share-0.2143) > 0.005 {
		t.Errorf("old server model share %.4f, want about 0.214", share)
	}
}
