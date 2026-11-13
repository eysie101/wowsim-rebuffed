package wotlk

import (
	"testing"

	"github.com/wowsims/wotlk/sim/core"
)

func TestLightweaveTrigger(t *testing.T) {
	rows := []struct {
		name  string
		spell core.Spell
		want  bool
	}{
		{"main-hand auto attack", core.Spell{ActionID: core.ActionID{OtherID: 3, Tag: 1}, SpellSchool: core.SpellSchoolPhysical, ProcMask: core.ProcMaskMeleeMHAuto}, false},
		{"off-hand auto attack", core.Spell{ActionID: core.ActionID{OtherID: 3, Tag: 2}, SpellSchool: core.SpellSchoolPhysical, ProcMask: core.ProcMaskMeleeOHAuto}, false},
		{"ranged auto attack", core.Spell{ActionID: core.ActionID{OtherID: 4}, SpellSchool: core.SpellSchoolPhysical, ProcMask: core.ProcMaskRangedAuto}, false},
		{"Stormstrike", core.Spell{ActionID: core.ActionID{SpellID: 17364}, SpellSchool: core.SpellSchoolPhysical, ProcMask: core.ProcMaskMeleeMHSpecial}, false},
		{"Lava Lash", core.Spell{ActionID: core.ActionID{SpellID: 60103}, SpellSchool: core.SpellSchoolFire, ProcMask: core.ProcMaskMeleeOHSpecial}, false},
		{"physical spell", core.Spell{ActionID: core.ActionID{SpellID: 1}, SpellSchool: core.SpellSchoolPhysical, ProcMask: core.ProcMaskEmpty}, false},
		{"Magma Totem", core.Spell{ActionID: core.ActionID{SpellID: 58734}, SpellSchool: core.SpellSchoolFire, ProcMask: core.ProcMaskEmpty}, false},
		{"Searing Totem", core.Spell{ActionID: core.ActionID{SpellID: 58704}, SpellSchool: core.SpellSchoolFire, ProcMask: core.ProcMaskEmpty}, false},
		{"Saronite Bomb", core.Spell{ActionID: core.ActionID{ItemID: 41119}, SpellSchool: core.SpellSchoolFire, ProcMask: core.ProcMaskEmpty}, false},
		{"Flametongue Attack", core.Spell{ActionID: core.ActionID{SpellID: 58790}, SpellSchool: core.SpellSchoolFire, ProcMask: core.ProcMaskWeaponProc}, true},
		{"Lightning Shield orb", core.Spell{ActionID: core.ActionID{SpellID: 49279}, SpellSchool: core.SpellSchoolNature, ProcMask: core.ProcMaskEmpty}, true},
		{"Lightning Bolt", core.Spell{ActionID: core.ActionID{SpellID: 49238}, SpellSchool: core.SpellSchoolNature, ProcMask: core.ProcMaskSpellDamage}, true},
		{"Flame Shock", core.Spell{ActionID: core.ActionID{SpellID: 49233}, SpellSchool: core.SpellSchoolFire, ProcMask: core.ProcMaskSpellDamage}, true},
		{"Global Thermal Sapper Charge", core.Spell{ActionID: core.ActionID{ItemID: 42641}, SpellSchool: core.SpellSchoolFire, ProcMask: core.ProcMaskEmpty}, true},
		{"Flamecap Fire", core.Spell{ActionID: core.ActionID{ItemID: 22788}, SpellSchool: core.SpellSchoolFire, ProcMask: core.ProcMaskEmpty}, true},
		{"Shadow Bolt", core.Spell{ActionID: core.ActionID{SpellID: 47809}, SpellSchool: core.SpellSchoolShadow, ProcMask: core.ProcMaskSpellDamage}, true},
	}
	for _, r := range rows {
		s := r.spell
		if got := lightweaveTrigger(&s); got != r.want {
			t.Errorf("%s: lightweaveTrigger = %v, want %v", r.name, got, r.want)
		}
	}
}
