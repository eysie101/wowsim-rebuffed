package enhancement

import (
	"math"
	"testing"

	googleProto "google.golang.org/protobuf/proto"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/core/stats"
	"github.com/wowsims/wotlk/sim/shaman"
)

func TestWhitemaneSpiritWolves(t *testing.T) {
	for _, c := range []struct {
		name      string
		glyph     bool
		icyTalons bool
		windfury  proto.TristateEffect
		ownAir    proto.AirTotem
		share     float64
		speed     float64
		heroic    bool
	}{
		{"full raid buffs", true, true, proto.TristateEffect_TristateEffectImproved, proto.AirTotem_WindfuryTotem, 0.61, 1.2, false},
		{"full raid buffs and Heroic Presence", true, true, proto.TristateEffect_TristateEffectImproved, proto.AirTotem_WindfuryTotem, 0.61, 1.2, true},
		{"full raid buffs without Glyph of Feral Spirit", false, true, proto.TristateEffect_TristateEffectImproved, proto.AirTotem_WindfuryTotem, 0.31, 1.2, false},
		{"Improved Windfury Totem without Icy Talons", true, false, proto.TristateEffect_TristateEffectImproved, proto.AirTotem_WindfuryTotem, 0.61, 1.2, false},
		{"Windfury Totem without Icy Talons", true, false, proto.TristateEffect_TristateEffectRegular, proto.AirTotem_WrathOfAirTotem, 0.61, 1.16, false},
		{"no melee haste buff", true, false, proto.TristateEffect_TristateEffectMissing, proto.AirTotem_WrathOfAirTotem, 0.61, 1, false},
	} {
		rsr := whitemaneRequest(1)
		if c.heroic {
			party := googleProto.Clone(rsr.Raid.Parties[0].Buffs).(*proto.PartyBuffs)
			party.HeroicPresence = true
			rsr.Raid.Parties[0].Buffs = party
		}
		player := rsr.Raid.Parties[0].Players[0]
		if !c.glyph {
			player.Glyphs = &proto.Glyphs{
				Major1: int32(proto.ShamanMajorGlyph_GlyphOfFireNova),
				Major2: int32(proto.ShamanMajorGlyph_GlyphOfFlametongueWeapon),
				Major3: int32(proto.ShamanMajorGlyph_GlyphOfStormstrike),
			}
		}
		options := googleProto.Clone(enhShamFTFT).(*proto.EnhancementShaman_Options)
		options.Totems.Air = c.ownAir
		player.Spec = &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{Options: options}}
		raidBuffs := googleProto.Clone(core.FullRaidBuffs).(*proto.RaidBuffs)
		raidBuffs.IcyTalons, raidBuffs.WindfuryTotem = c.icyTalons, c.windfury
		rsr.Raid.Buffs = raidBuffs

		sim := core.NewSim(rsr)
		sim.Reset()
		enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
		enh.SpiritWolves.EnableWithTimeout(sim)
		wolves := []*shaman.SpiritWolf{enh.SpiritWolves.SpiritWolf1, enh.SpiritWolves.SpiritWolf2}

		for _, wolf := range wolves {
			base := wolf.GetBaseStats()
			if str, agi := wolf.GetStat(stats.Strength), wolf.GetStat(stats.Agility); math.Abs(str-base[stats.Strength]) > 1e-9 || math.Abs(agi-base[stats.Agility]) > 1e-9 {
				t.Errorf("%s, %s: Strength %.1f and Agility %.1f, want the base %.0f and %.0f without Strength of Earth, Blessing of Kings or Mark of the Wild", c.name, wolf.Name, str, agi, base[stats.Strength], base[stats.Agility])
			}
			if got, want := wolf.GetStat(stats.MeleeCrit), base[stats.MeleeCrit]+base[stats.Agility]*core.CritRatingPerCritChance/83.3; math.Abs(got-want) > 1e-6 {
				t.Errorf("%s, %s: melee crit rating %.2f, want %.2f from the base and Agility without Leader of the Pack", c.name, wolf.Name, got, want)
			}
			for _, label := range []string{"Battle Shout", "Blessing of Might"} {
				if aura := wolf.GetAura(label); aura != nil && aura.IsActive() {
					t.Errorf("%s, %s: %s is active on the wolf", c.name, wolf.Name, label)
				}
			}
			if got := wolf.PseudoStats.DamageDealtMultiplier; math.Abs(got-1.05) > 1e-9 {
				t.Errorf("%s, %s: damage dealt x%.4f, want x1.05 from the orc owner's Command alone, without Ferocious Inspiration, Arcane Empowerment or Sanctified Retribution", c.name, wolf.Name, got)
			}
			if got := wolf.SwingSpeed(); math.Abs(got-c.speed) > 1e-9 {
				t.Errorf("%s, %s: attack speed x%.4f, want x%.2f from the raid's melee haste buff alone", c.name, wolf.Name, got, c.speed)
			}
			ownerHit := math.Floor(enh.GetStat(stats.MeleeHit)/core.MeleeHitRatingPerHitChance) * core.MeleeHitRatingPerHitChance
			if got, want := wolf.GetStat(stats.MeleeHit), base[stats.MeleeHit]+ownerHit; math.Abs(got-want) > 1e-6 {
				t.Errorf("%s, %s: melee hit rating %.2f, want %.2f from the base and the owner's whole hit percent, without a hit buff of its own", c.name, wolf.Name, got, want)
			}
		}

		checkAttackPower := func(when string) {
			ownerAP := enh.GetStat(stats.AttackPower)
			for _, wolf := range wolves {
				if got, want := wolf.GetStat(stats.AttackPower), 311+c.share*ownerAP; math.Abs(got-want) > 1e-6 {
					t.Errorf("%s, %s %s: attack power %.2f, want 311 + %.2f x owner attack power %.2f = %.2f", c.name, wolf.Name, when, got, c.share, ownerAP, want)
				}
			}
		}
		checkAttackPower("at the summon")
		enh.AddStatDynamic(sim, stats.AttackPower, 1000)
		checkAttackPower("after the owner gained 1000 attack power")
	}
}

func TestWhitemaneSpiritWolvesGetOwnersBloodlust(t *testing.T) {
	sim := core.NewSim(whitemaneRequest(1))
	sim.Reset()
	enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
	enh.SpiritWolves.EnableWithTimeout(sim)
	wolves := []*shaman.SpiritWolf{enh.SpiritWolves.SpiritWolf1, enh.SpiritWolves.SpiritWolf2}
	before := make([]float64, len(wolves))
	for i, wolf := range wolves {
		before[i] = wolf.SwingSpeed()
	}
	bloodlust := enh.GetAuraByID(core.BloodlustActionID.WithTag(-1))
	if bloodlust == nil {
		t.Fatal("the shaman has no Bloodlust aura")
	}
	bloodlust.Activate(sim)
	for i, wolf := range wolves {
		active := false
		for _, aura := range wolf.GetAuras() {
			if aura.Tag == core.BloodlustAuraTag && aura.IsActive() {
				active = true
			}
		}
		if ratio := wolf.SwingSpeed() / before[i]; !active || math.Abs(ratio-1.3) > 1e-9 {
			t.Errorf("%s: Bloodlust active %v, attack speed x%.4f -> x%.4f, want the owner's Bloodlust at x1.3", wolf.Name, active, before[i], wolf.SwingSpeed())
		}
	}
}
