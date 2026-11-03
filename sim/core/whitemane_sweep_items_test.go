package core

import (
	"testing"

	"github.com/wowsims/wotlk/sim/core/proto"
)

func whitemaneSweepArmor(raidBuffs *proto.RaidBuffs) float64 {
	raid := SinglePlayerRaidProto(&proto.Player{
		Name:      "Caster",
		Race:      proto.Race_RaceTroll,
		Class:     proto.Class_ClassShaman,
		Spec:      &proto.Player_ElementalShaman{},
		Consumes:  &proto.Consumes{},
		Buffs:     &proto.IndividualBuffs{},
		Equipment: &proto.EquipmentSpec{},
	}, &proto.PartyBuffs{}, raidBuffs, &proto.Debuffs{})
	encounter := &proto.Encounter{Duration: 60, Targets: []*proto.Target{NewDefaultTarget()}}
	_, raidStats, _ := NewEnvironment(raid, encounter, false)
	return raidStats.Parties[0].Players[0].FinalStats.Stats[proto.Stat_StatArmor]
}

func TestWhitemaneSweepArmorScrollAndMarkOfTheWild(t *testing.T) {
	unbuffed := whitemaneSweepArmor(&proto.RaidBuffs{})
	if got := whitemaneSweepArmor(&proto.RaidBuffs{ScrollOfProtection: true}) - unbuffed; got != 750 {
		t.Errorf("Scroll of Protection alone adds %.0f armor, want 750", got)
	}
	if got := whitemaneSweepArmor(&proto.RaidBuffs{ScrollOfProtection: true, DevotionAura: proto.TristateEffect_TristateEffectRegular}) -
		whitemaneSweepArmor(&proto.RaidBuffs{DevotionAura: proto.TristateEffect_TristateEffectRegular}); got != 0 {
		t.Errorf("Scroll of Protection adds %.0f armor on top of Devotion Aura, want 0", got)
	}
	for _, c := range []struct {
		name  string
		buffs func() *proto.RaidBuffs
	}{
		{"Mark of the Wild", func() *proto.RaidBuffs {
			return &proto.RaidBuffs{GiftOfTheWild: proto.TristateEffect_TristateEffectRegular}
		}},
		{"Improved Mark of the Wild", func() *proto.RaidBuffs {
			return &proto.RaidBuffs{GiftOfTheWild: proto.TristateEffect_TristateEffectImproved}
		}},
		{"Drums of the Wild", func() *proto.RaidBuffs { return &proto.RaidBuffs{DrumsOfTheWild: true} }},
	} {
		without := whitemaneSweepArmor(c.buffs())
		withScroll := c.buffs()
		withScroll.ScrollOfProtection = true
		got := whitemaneSweepArmor(withScroll) - without
		t.Logf("%s: %.0f armor, %+.0f with Scroll of Protection", c.name, without-unbuffed, got)
		if got != 0 {
			t.Errorf("Scroll of Protection adds %.0f armor on top of %s, want 0", got, c.name)
		}
	}
}
