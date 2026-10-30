package enhancement

import (
	"testing"

	"github.com/wowsims/wotlk/sim/core"
)

func whitemaneJudgementOfWisdomSim(jarAndIcebreaker bool) (*core.Simulation, *EnhancementShaman) {
	rsr := whitemaneRequest(1)
	player := rsr.Raid.Parties[0].Players[0]
	player.Spec = PlayerOptionsWFWF
	player.Rotation = core.GetAplRotation("../../../ui/enhancement_shaman/apls", "enha_p1_wf").Rotation
	if jarAndIcebreaker {
		player.Equipment.Items[12].Id = 50351
		player.Equipment.Items[14].Enchant = 3239
	}
	sim := core.NewSim(rsr)
	return sim, sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
}

func TestWhitemaneJudgementOfWisdomEnhancementFight(t *testing.T) {
	sim, enh := whitemaneJudgementOfWisdomSim(false)
	wfMH := enh.GetSpell(core.ActionID{SpellID: 58804, Tag: 1})
	wfOH := enh.GetSpell(core.ActionID{SpellID: 58804, Tag: 2})
	kind := func(spell *core.Spell) string {
		switch spell {
		case wfMH, wfOH:
			return "Windfury Attack"
		case enh.Stormstrike:
			return "Stormstrike strike"
		case enh.AutoAttacks.MHAuto(), enh.AutoAttacks.OHAuto():
			return "melee swing"
		}
		return "player cast"
	}
	hits, procs := map[string]int{}, map[string]int{}
	jow := sim.Encounter.TargetUnits[0].GetAura(core.JudgementOfWisdomAuraLabel)
	handler := jow.OnSpellHitTaken
	jow.OnSpellHitTaken = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		events := int32(0)
		if enh.JowManaMetrics != nil {
			events = enh.JowManaMetrics.Events
		}
		handler(aura, sim, spell, result)
		if spell.Unit != &enh.Unit {
			return
		}
		hits[kind(spell)]++
		if enh.JowManaMetrics != nil && enh.JowManaMetrics.Events > events {
			procs[kind(spell)]++
		}
	}
	sim.Reset()
	sim.PrePull()
	for !sim.Step() {
	}
	sim.Cleanup()

	if enh.JowManaMetrics == nil {
		t.Fatalf("no Judgement of Wisdom mana in the fight; hits %v", hits)
	}
	t.Logf("hits %v, Judgement of Wisdom procs %v, %.0f mana from %d procs", hits, procs, enh.JowManaMetrics.Gain, enh.JowManaMetrics.Events)
	for _, k := range []string{"Windfury Attack", "Stormstrike strike"} {
		if hits[k] < 20 {
			t.Fatalf("%d %s hits in the fight; the check needs at least 20", hits[k], k)
		}
		if procs[k] != 0 {
			t.Errorf("%d of %d %s hits procced Judgement of Wisdom, want none", procs[k], hits[k], k)
		}
	}
	for _, k := range []string{"melee swing", "player cast"} {
		if procs[k] == 0 {
			t.Errorf("none of %d %s hits procced Judgement of Wisdom", hits[k], k)
		}
	}
}

func TestWhitemaneJudgementOfWisdomEnhancementSpells(t *testing.T) {
	sim, enh := whitemaneJudgementOfWisdomSim(true)
	sim.Reset()
	for _, c := range []struct {
		name  string
		spell *core.Spell
		procs bool
	}{
		{"Windfury Attack main hand", enh.GetSpell(core.ActionID{SpellID: 58804, Tag: 1}), false},
		{"Windfury Attack off hand", enh.GetSpell(core.ActionID{SpellID: 58804, Tag: 2}), false},
		{"Stormstrike", enh.Stormstrike, false},
		{"Fire Nova", enh.FireNova, false},
		{"Manifest Anger main hand", enh.GetSpell(core.ActionID{SpellID: 71433}), false},
		{"Manifest Anger off hand", enh.GetSpell(core.ActionID{SpellID: 71434}), false},
		{"Icebreaker", enh.GetSpell(core.ActionID{SpellID: 44525}), false},
		{"main-hand swing", enh.AutoAttacks.MHAuto(), true},
		{"Lava Lash", enh.LavaLash, true},
		{"Earth Shock", enh.EarthShock, true},
	} {
		if c.spell == nil {
			t.Fatalf("%s not registered", c.name)
		}
		if got := core.JudgementOfWisdomManaEvents(sim, c.spell, 200); (got > 0) != c.procs {
			t.Errorf("%s: %d Judgement of Wisdom procs from 200 hits, want procs %v", c.name, got, c.procs)
		}
	}
}
