package hunter

import (
	"time"

	"github.com/wowsims/wotlk/sim/core"
)

func (hunter *Hunter) registerKillCommandCD() {
	if hunter.pet == nil {
		return
	}

	actionID := core.ActionID{SpellID: 34026}
	bonusPetSpecialCrit := 10 * core.CritRatingPerCritChance * float64(hunter.Talents.FocusedFire)
	specialAttacks := core.ProcMaskMeleeSpecial | core.ProcMaskSpellDamage

	var petSpecials []*core.Spell
	hunter.pet.OnSpellRegistered(func(spell *core.Spell) {
		if spell.ProcMask.Matches(specialAttacks) {
			petSpecials = append(petSpecials, spell)
		}
	})

	hunter.pet.KillCommandAura = hunter.pet.RegisterAura(core.Aura{
		Label:     "Kill Command",
		ActionID:  actionID,
		Duration:  time.Second * 30,
		MaxStacks: 3,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range petSpecials {
				spell.BonusCritRating += bonusPetSpecialCrit
			}
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			for _, spell := range petSpecials {
				spell.BonusCritRating -= bonusPetSpecialCrit
			}
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.ProcMask.Matches(specialAttacks) {
				aura.RemoveStack(sim)
			}
		},
	})

	hunter.KillCommand = hunter.RegisterSpell(core.SpellConfig{
		ActionID:    actionID,
		SpellSchool: core.SpellSchoolPhysical,
		Flags:       core.SpellFlagNoOnCastComplete,

		ManaCost: core.ManaCostOptions{
			BaseCost: 0.03,
		},
		Cast: core.CastConfig{
			CD: core.Cooldown{
				Timer:    hunter.NewTimer(),
				Duration: time.Minute - time.Second*10*time.Duration(hunter.Talents.CatlikeReflexes),
			},
		},
		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return hunter.pet.IsEnabled()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			hunter.pet.KillCommandAura.Activate(sim)
			hunter.pet.KillCommandAura.SetStacks(sim, 3)
		},
	})

	hunter.AddMajorCooldown(core.MajorCooldown{
		Spell: hunter.KillCommand,
		Type:  core.CooldownTypeDPS,
	})
}
