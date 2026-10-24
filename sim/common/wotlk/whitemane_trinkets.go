package wotlk

import (
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/stats"
)

func init() {
	core.AddEffectsToTest = false

	onUse := func(itemID int32, bonus stats.Stats, duration time.Duration) {
		core.NewSimpleStatOffensiveTrinketEffect(itemID, bonus, duration, time.Minute*2)
	}
	haste := func(v float64) stats.Stats { return stats.Stats{stats.MeleeHaste: v, stats.SpellHaste: v} }
	attackPower := func(ap, rap float64) stats.Stats {
		return stats.Stats{stats.AttackPower: ap, stats.RangedAttackPower: rap}
	}
	spellPower := func(v float64) stats.Stats { return stats.Stats{stats.SpellPower: v} }

	onUse(1207310, haste(194), time.Second*20)
	onUse(1207314, spellPower(120), time.Second*15)
	onUse(1207316, attackPower(346, 264), time.Second*20)
	onUse(1207318, attackPower(367, 280), time.Second*20)
	onUse(1207700, spellPower(145), time.Second*20)
	onUse(1207702, attackPower(284, 248), time.Second*20)
	onUse(1207704, spellPower(145), time.Second*20)
	onUse(1207708, haste(239), time.Second*20)
	onUse(1207710, spellPower(202), time.Second*20)
	onUse(1208214, spellPower(183), time.Second*20)
	onUse(1208230, spellPower(145), time.Second*20)
	onUse(1208911, haste(243), time.Second*20)
	onUse(1208913, spellPower(281), time.Second*20)

	core.NewSimpleStatDefensiveTrinketEffect(1207706, stats.Stats{stats.Spirit: 211}, time.Second*20, time.Minute*2)
	core.NewSimpleStatDefensiveTrinketEffect(1209272, stats.Stats{stats.Dodge: 292}, time.Second*20, time.Minute*2)

	newProcStatBonusEffect(ProcStatBonusEffect{
		Name:       "Valonforth's Remembrance",
		ID:         1207312,
		AuraID:     9250494,
		Bonus:      stats.Stats{stats.Spirit: 291},
		Duration:   time.Second * 10,
		Callback:   core.CallbackOnSpellHitDealt | core.CallbackOnHealDealt,
		ProcMask:   core.ProcMaskSpellDamage | core.ProcMaskSpellHealing,
		Outcome:    core.OutcomeLanded,
		ProcChance: 0.15,
		ICD:        time.Second * 45,
	})
	newProcStatBonusEffect(ProcStatBonusEffect{
		Name:       "Crusader's Locket",
		ID:         1208897,
		AuraID:     9250553,
		Bonus:      stats.Stats{stats.Expertise: 296},
		Duration:   time.Second * 10,
		Callback:   core.CallbackOnSpellHitDealt,
		ProcMask:   core.ProcMaskMeleeOrRanged,
		Outcome:    core.OutcomeLanded,
		ProcChance: 0.15,
		ICD:        time.Second * 45,
	})
	newProcStatBonusEffect(ProcStatBonusEffect{
		Name:       "Chuchu's Tiny Box of Horrors",
		ID:         1208915,
		AuraID:     9250561,
		Bonus:      stats.Stats{stats.MeleeCrit: 296, stats.SpellCrit: 296},
		Duration:   time.Second * 10,
		Callback:   core.CallbackOnSpellHitDealt,
		ProcMask:   core.ProcMaskMeleeOrRanged,
		Outcome:    core.OutcomeLanded,
		ProcChance: 0.15,
		ICD:        time.Second * 45,
	})
	newProcStatBonusEffect(ProcStatBonusEffect{
		Name:       "Coren's Chromium Coaster",
		ID:         1209269,
		AuraID:     9250856,
		Bonus:      stats.Stats{stats.AttackPower: 873, stats.RangedAttackPower: 873},
		Duration:   time.Second * 10,
		Callback:   core.CallbackOnSpellHitDealt,
		ProcMask:   core.ProcMaskMeleeOrRanged,
		Outcome:    core.OutcomeCrit,
		ProcChance: 0.10,
		ICD:        time.Second * 45,
	})
	newProcStatBonusEffect(ProcStatBonusEffect{
		Name:       "Mithril Pocketwatch",
		ID:         1209270,
		AuraID:     9250858,
		Bonus:      stats.Stats{stats.SpellPower: 515},
		Duration:   time.Second * 10,
		Callback:   core.CallbackOnSpellHitDealt,
		ProcMask:   core.ProcMaskSpellOrProc,
		Harmful:    true,
		ProcChance: 0.10,
		ICD:        time.Second * 45,
	})
	newProcStatBonusEffect(ProcStatBonusEffect{
		Name:       "Ancient Pickled Egg",
		ID:         1209271,
		AuraID:     9250860,
		Bonus:      stats.Stats{stats.MeleeHaste: 441, stats.SpellHaste: 441},
		Duration:   time.Second * 10,
		Callback:   core.CallbackOnHealDealt | core.CallbackOnPeriodicHealDealt,
		ProcChance: 0.10,
		ICD:        time.Second * 45,
	})
	newProcDamageEffect(ProcDamageEffect{
		ID: 1208352,
		Trigger: core.ProcTrigger{
			Name:       "Horn of Argent Fury",
			Callback:   core.CallbackOnSpellHitDealt,
			ProcMask:   core.ProcMaskMeleeOrRanged,
			Outcome:    core.OutcomeLanded,
			ProcChance: 0.15,
			ICD:        time.Second * 45,
			ActionID:   core.ActionID{ItemID: 1208352},
		},
		School: core.SpellSchoolHoly,
		MinDmg: 1184,
		MaxDmg: 1776,
	})

	core.AddEffectsToTest = true
}
