package core

import (
	"math"
	"testing"

	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/core/stats"
)

func init() {
	RegisterAgentFactory(
		proto.Player_Mage{},
		proto.Spec_SpecMage,
		newWhitemaneSweepOwner,
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_Mage)
		},
	)
}

type whitemaneSweepOwner struct {
	Character
	pets []*whitemaneSweepPet
}

func (owner *whitemaneSweepOwner) GetCharacter() *Character { return &owner.Character }
func (owner *whitemaneSweepOwner) Initialize()              {}
func (owner *whitemaneSweepOwner) ApplyTalents()            {}
func (owner *whitemaneSweepOwner) Reset(_ *Simulation)      {}

type whitemaneSweepPet struct {
	Pet
	want float64
}

func (pet *whitemaneSweepPet) GetPet() *Pet                        { return &pet.Pet }
func (pet *whitemaneSweepPet) Initialize()                         {}
func (pet *whitemaneSweepPet) Reset(_ *Simulation)                 {}
func (pet *whitemaneSweepPet) ExecuteCustomRotation(_ *Simulation) {}

func newWhitemaneSweepOwner(character *Character, _ *proto.Player) Agent {
	owner := &whitemaneSweepOwner{Character: *character}
	noHit := func(stats.Stats) stats.Stats { return stats.Stats{} }
	ownerHit := func(ownerStats stats.Stats) stats.Stats {
		return stats.Stats{stats.MeleeHit: ownerStats[stats.MeleeHit], stats.SpellHit: ownerStats[stats.SpellHit]}
	}
	for _, pet := range []*whitemaneSweepPet{
		{Pet: NewPet("Pet", &owner.Character, stats.Stats{}, noHit, true, false), want: 1},
		{Pet: NewPet("Hit Pet", &owner.Character, stats.Stats{}, ownerHit, true, false), want: 2},
		{Pet: NewPet("Guardian", &owner.Character, stats.Stats{}, noHit, false, true), want: 1},
		{Pet: NewPet("Hit Guardian", &owner.Character, stats.Stats{}, ownerHit, false, true), want: 1},
	} {
		owner.pets = append(owner.pets, pet)
		owner.AddPet(pet)
	}
	return owner
}

func whitemaneSweepHeroicPresenceSim(source string) *Simulation {
	player := func(race proto.Race) *proto.Player {
		return &proto.Player{
			Name:      "Mage",
			Race:      race,
			Class:     proto.Class_ClassMage,
			Spec:      &proto.Player_Mage{},
			Consumes:  &proto.Consumes{},
			Buffs:     &proto.IndividualBuffs{},
			Equipment: &proto.EquipmentSpec{},
		}
	}
	first, firstBuffs := player(proto.Race_RaceHuman), &proto.PartyBuffs{}
	switch source {
	case "Draenei":
		first.Race = proto.Race_RaceDraenei
	case "party buff":
		firstBuffs.HeroicPresence = true
	}
	sim := NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{
			Parties: []*proto.Party{
				{Players: []*proto.Player{first}, Buffs: firstBuffs},
				{Players: []*proto.Player{player(proto.Race_RaceHuman)}, Buffs: &proto.PartyBuffs{}},
			},
			Buffs: &proto.RaidBuffs{MoonkinAura: proto.TristateEffect_TristateEffectRegular},
		},
		Encounter: &proto.Encounter{Duration: 30, Targets: []*proto.Target{{Name: "target", Level: 83}}},
	})
	sim.Reset()
	for _, party := range sim.Raid.Parties {
		for _, pet := range party.Players[0].(*whitemaneSweepOwner).pets {
			if !pet.IsEnabled() {
				pet.Enable(sim, pet)
			}
		}
	}
	return sim
}

func TestWhitemaneSweepHeroicPresenceRaidPetsAndSummons(t *testing.T) {
	control := whitemaneSweepHeroicPresenceSim("")
	for _, source := range []string{"Draenei", "party buff"} {
		sim := whitemaneSweepHeroicPresenceSim(source)
		for partyIdx, party := range sim.Raid.Parties {
			owner := party.Players[0].(*whitemaneSweepOwner)
			controlOwner := control.Raid.Parties[partyIdx].Players[0].(*whitemaneSweepOwner)
			units := []*Character{&owner.Character}
			controlUnits := []*Character{&controlOwner.Character}
			wants := []float64{1}
			for i, pet := range owner.pets {
				units = append(units, &pet.Character)
				controlUnits = append(controlUnits, &controlOwner.pets[i].Character)
				wants = append(wants, pet.want)
			}
			for i, unit := range units {
				melee := (unit.GetStat(stats.MeleeHit) - controlUnits[i].GetStat(stats.MeleeHit)) / MeleeHitRatingPerHitChance
				spell := (unit.GetStat(stats.SpellHit) - controlUnits[i].GetStat(stats.SpellHit)) / SpellHitRatingPerHitChance
				t.Logf("source %s, party %d, %s: %+.2f%% melee hit, %+.2f%% spell hit", source, partyIdx+1, unit.Label, melee, spell)
				if math.Abs(melee-wants[i]) > 1e-9 || math.Abs(spell-wants[i]) > 1e-9 {
					t.Errorf("source %s, party %d, %s: Heroic Presence adds %.2f%% melee and %.2f%% spell hit, want %.0f%% each", source, partyIdx+1, unit.Label, melee, spell, wants[i])
				}
			}
			for _, pet := range owner.pets {
				if crit := pet.GetStat(stats.SpellCrit); pet.IsGuardian() == (crit != 0) {
					t.Errorf("source %s, party %d, %s: %.0f spell crit rating from Moonkin Aura, a guardian gets no raid buff but Heroic Presence and a pet gets all", source, partyIdx+1, pet.Label, crit)
				}
			}
		}
	}
}
