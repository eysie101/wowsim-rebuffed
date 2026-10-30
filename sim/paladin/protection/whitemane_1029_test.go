package protection

import (
	"testing"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func TestWhitemaneArdentDefenderAfterAbsorbs(t *testing.T) {
	rsr := &proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Prot",
				Race:          proto.Race_RaceBloodElf,
				Class:         proto.Class_ClassPaladin,
				Equipment:     core.GetGearSet("../../../ui/protection_paladin/gear_sets", "p1").GearSet,
				TalentsString: StandardTalents,
				Glyphs:        StandardGlyphs,
				Spec:          DefaultOptions,
				Buffs:         &proto.IndividualBuffs{GuardianSpirits: 1},
			},
			&proto.PartyBuffs{},
			&proto.RaidBuffs{},
			&proto.Debuffs{}),
		Encounter: &proto.Encounter{
			Duration: 60,
			Targets:  []*proto.Target{core.NewDefaultTarget()},
		},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1, IsTest: true},
	}
	rsr.Raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}

	sim := core.NewSim(rsr)
	sim.Reset()
	prot, ok := sim.Raid.Parties[0].Players[0].(*ProtectionPaladin)
	if !ok {
		t.Fatalf("player is %T, not a Protection paladin", sim.Raid.Parties[0].Players[0])
	}
	if prot.Talents.ArdentDefender != 3 {
		t.Fatalf("Ardent Defender %d/3 in the test talents, want 3/3", prot.Talents.ArdentDefender)
	}
	guardianSpirit := prot.GetAuraByID(core.ActionID{SpellID: 47788, Tag: -1})
	ardentDefender := prot.GetAura("Ardent Defender")
	if guardianSpirit == nil || ardentDefender == nil {
		t.Fatal("Guardian Spirit or Ardent Defender aura not registered")
	}

	prot.RemoveHealth(sim, 0.5*prot.MaxHealth())
	guardianSpirit.Activate(sim)
	boss := sim.Encounter.Targets[0].AutoAttacks.MHAuto()
	boss.CalcAndDealDamage(sim, &prot.Unit, 10*prot.MaxHealth(), boss.OutcomeAlwaysHit)

	if guardianSpirit.IsActive() {
		t.Errorf("Guardian Spirit is still up after a killing blow: Ardent Defender took the blow before the absorb")
	}
	if ardentDefender.IsActive() {
		t.Errorf("Ardent Defender's heal fired although Guardian Spirit absorbs the killing blow first")
	}
}
