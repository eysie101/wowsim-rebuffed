package rogue

import (
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/stats"
)

var Arena = core.NewItemSet(core.ItemSet{
	Name: "Gladiator's Vestments",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.Resilience, 100)
			agent.GetCharacter().AddStat(stats.AttackPower, 50)
		},
		4: func(agent core.Agent) {
			agent.GetCharacter().AddStat(stats.AttackPower, 150)
			// 10 maximum energy added in rogue.go
		},
	},
})

var Tier10 = core.NewItemSet(core.ItemSet{
	Name: "Shadowblade's Battlegear",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			// Your Tricks of the Trade now grants you 15 energy instead of costing energy.
			// Handled in tricks_of_the_trade.go.
		},
		4: func(agent core.Agent) {
			// Gives your melee finishing moves a 13% chance to add 3 combo points to your target.
			// Handled in the finishing move effect applier
		},
	},
})

var Tier9 = core.NewItemSet(core.ItemSet{
	Name:            "VanCleef's Battlegear",
	AlternativeName: "Garona's Battlegear",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			// Your Rupture ability has a chance each time it deals damage to reduce the cost of your next ability by 40 energy.
			rogue := agent.(RogueAgent).GetRogue()
			energyMetrics := rogue.NewEnergyMetrics(core.ActionID{SpellID: 67209})

			procAura := rogue.RegisterAura(core.Aura{
				Label:    "VanCleef's 2pc Proc",
				ActionID: core.ActionID{SpellID: 67209},
				Duration: time.Second * 15,
				OnGain: func(aura *core.Aura, sim *core.Simulation) {
					rogue.PseudoStats.CostReduction += 40
				},
				OnExpire: func(aura *core.Aura, sim *core.Simulation) {
					rogue.PseudoStats.CostReduction -= 40
				},
				OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					if !spell.ProcMask.Matches(core.ProcMaskMeleeSpecial) {
						return
					}

					// doesn't handle multiple dynamic cost reductions at once, or 0-cost default casts
					if actualGain := spell.DefaultCast.Cost - spell.CurCast.Cost; actualGain > 0 {
						energyMetrics.AddEvent(40, actualGain)
						aura.Deactivate(sim)
					}
				},
			})

			icd := core.Cooldown{
				Timer:    rogue.NewTimer(),
				Duration: time.Second * 15,
			}
			procAura.Icd = &icd
			procChance := 0.02
			rogue.RegisterAura(core.Aura{
				Label:    "VanCleef's 2pc",
				Duration: core.NeverExpires,
				OnReset: func(aura *core.Aura, sim *core.Simulation) {
					aura.Activate(sim)
				},
				OnPeriodicDamageDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if !result.Landed() {
						return
					}
					if !spell.ActionID.IsSpellAction(RuptureSpellID) {
						return
					}
					if !icd.IsReady(sim) {
						return
					}
					if sim.RandomFloat("VanCleef's 2pc") > procChance {
						return
					}
					icd.Use(sim)
					procAura.Activate(sim)
				},
			})
		},
		4: func(agent core.Agent) {
			// Increases the critical strike chance of your Hemorrhage, Sinister Strike, Backstab, and Mutilate abilities by 5%.
			// Handled in ability sources
		},
	},
})

var Tier8 = core.NewItemSet(core.ItemSet{
	Name: "Terrorblade Battlegear",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			// Your Deadly Poison causes you to gain 1 energy each time it deals damage
			// Handled in poisons.go
		},
		4: func(agent core.Agent) {
			// Increases the damage done by your Rupture by 20%
			// Handled in rupture.go
		},
	},
})

var Tier7 = core.NewItemSet(core.ItemSet{
	Name: "Bonescythe Battlegear",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			// Increases the damage dealt by your Rupture by 10%
			// Handled in rupture.go
		},
		4: func(agent core.Agent) {
			// Reduce the Energy cost of your Combo Moves by 5%
			// Handled in the builder cast modifier
		},
	},
})

var Tier6 = core.NewItemSet(core.ItemSet{
	Name: "Slayer's Armor",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			// Increases the haste from your Slice and Dice ability by 5%.
			// Handled in slice_and_dice.go.
		},
		4: func(agent core.Agent) {
			// Increases the damage dealt by your Backstab, Sinister Strike, Mutilate, and Hemorrhage abilities by 6%.
			// Handled in the corresponding ability files.
		},
	},
})

// Whitemane Warglaives
var ItemSetWarglaivesWhitemane = core.NewItemSet(core.ItemSet{
	Name: "Warglaives (Whitemane)",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			// Warglaives on Whitemane. copy-paste of Bandit's Guile from Cata
			character := (agent.(RogueAgent).GetRogue())
			attackCounter := int32(0)
			var lastAttacked *core.Unit
			var bgDamageAuras [3]*core.Aura
			currentInsightIndex := -1

			for index := 0; index < 3; index++ {
				var label string
				var actionID core.ActionID
				switch index {
				case 0:
					label = "Shallow Insight"
					actionID = core.ActionID{SpellID: 54262}
				case 1:
					label = "Moderate Insight"
					actionID = core.ActionID{SpellID: 45797}
				case 2:
					label = "Deep Insight"
					actionID = core.ActionID{SpellID: 52679}
				}

				damageBonus := []float64{1.03, 1.06, 1.1}[index]

				bgDamageAuras[index] = character.RegisterAura(core.Aura{
					Label:    label,
					ActionID: actionID,
					Duration: time.Second * 15,

					OnGain: func(aura *core.Aura, sim *core.Simulation) {
						character.PseudoStats.DamageDealtMultiplier *= damageBonus
					},
					OnExpire: func(aura *core.Aura, sim *core.Simulation) {
						character.PseudoStats.DamageDealtMultiplier /= damageBonus
						if currentInsightIndex == 2 {
							currentInsightIndex = -1
							attackCounter = 0
						}
					},
				})
			}

			character.WarglaivesAura = character.RegisterAura(core.Aura{
				Label:     "Warglaives Tracker",
				ActionID:  core.ActionID{SpellID: 31284},
				Duration:  core.NeverExpires,
				MaxStacks: 4,
				OnReset: func(aura *core.Aura, sim *core.Simulation) {
					currentInsightIndex = -1
					attackCounter = 0
				},
				OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
					if currentInsightIndex < 2 && result.Landed() && (spell == (agent.(RogueAgent).GetRogue()).SinisterStrike) {
						if lastAttacked != result.Target {
							// Reset back to no insight, no casts
							attackCounter = 0
							if currentInsightIndex >= 0 {
								bgDamageAuras[currentInsightIndex].Deactivate(sim)
							}
							currentInsightIndex = -1
						}
						lastAttacked = result.Target

						attackCounter += 1
						character.WarglaivesAura.SetStacks(sim, attackCounter+1)
						if attackCounter == 4 {
							attackCounter = 0
							character.WarglaivesAura.SetStacks(sim, attackCounter+1)
							// Deactivate previous aura
							if currentInsightIndex >= 0 {
								bgDamageAuras[currentInsightIndex].Deactivate(sim)
							}
							currentInsightIndex += 1
							// Activate next aura
							bgDamageAuras[currentInsightIndex].Activate(sim)
						} else {
							// Refresh duration of existing aura
							if currentInsightIndex >= 0 {
								bgDamageAuras[currentInsightIndex].Duration = time.Second * 15
								bgDamageAuras[currentInsightIndex].Activate(sim)
							}
						}

					}
				},
			})
		},
	},
})

func init() {
	core.NewItemEffect(32492, func(agent core.Agent) {
		rogue := agent.(RogueAgent).GetRogue()
		procAura := rogue.NewTemporaryStatsAura("Ashtongue Talisman Proc", core.ActionID{ItemID: 32492}, stats.Stats{stats.MeleeCrit: 145}, time.Second*10)

		var numPoints int32

		rogue.RegisterAura(core.Aura{
			Label:    "Ashtongue Talisman",
			Duration: core.NeverExpires,
			OnReset: func(aura *core.Aura, sim *core.Simulation) {
				numPoints = 0
				aura.Activate(sim)
			},
			OnCastComplete: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
				if !spell.Flags.Matches(SpellFlagFinisher) {
					return
				}

				// Need to store the points because they get spent before OnSpellHit is called.
				numPoints = rogue.ComboPoints()

				if spell == rogue.SliceAndDice {
					// SND won't call OnSpellHit, so we have to add the effect now.
					if p := 0.2 * float64(numPoints); sim.Proc(p, "AshtongueTalismanOfLethality") {
						procAura.Activate(sim)
					}
				}
			},
			OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				if !spell.Flags.Matches(SpellFlagFinisher) {
					return
				}

				if p := 0.2 * float64(numPoints); sim.Proc(p, "AshtongueTalismanOfLethality") {
					procAura.Activate(sim)
				}
			},
		})
	})

}
