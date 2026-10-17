import { IMessageType } from '@protobuf-ts/runtime';

import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs, TristateEffect } from '../proto/common';

// The buffs and debuffs of a full 25-man raid: the sim's own full set (FullRaidBuffs, FullPartyBuffs,
// FullIndividualBuffs and FullDebuffs in sim/core/test_utils.go) plus Heroic Presence.
// Every spec starts with these, on top of its own defaults.
export const FULL_RAID_BUFFS = RaidBuffs.create({
	abominationsMight: true,
	arcaneBrilliance: true,
	arcaneEmpowerment: true,
	battleShout: TristateEffect.TristateEffectImproved,
	bloodlust: true,
	demonicPactSp: 500,
	devotionAura: TristateEffect.TristateEffectImproved,
	divineSpirit: true,
	elementalOath: true,
	ferociousInspiration: true,
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	icyTalons: true,
	leaderOfThePack: TristateEffect.TristateEffectImproved,
	manaSpringTotem: TristateEffect.TristateEffectRegular,
	moonkinAura: TristateEffect.TristateEffectImproved,
	powerWordFortitude: TristateEffect.TristateEffectImproved,
	sanctifiedRetribution: true,
	shadowProtection: true,
	strengthOfEarthTotem: TristateEffect.TristateEffectImproved,
	swiftRetribution: true,
	thorns: TristateEffect.TristateEffectImproved,
	totemOfWrath: true,
	trueshotAura: true,
	unleashedRage: true,
	windfuryTotem: TristateEffect.TristateEffectImproved,
	wrathOfAirTotem: true,
});

export const FULL_PARTY_BUFFS = PartyBuffs.create({
	manaTideTotems: 1,
	heroicPresence: true,
});

export const FULL_INDIVIDUAL_BUFFS = IndividualBuffs.create({
	blessingOfKings: true,
	blessingOfMight: TristateEffect.TristateEffectImproved,
	blessingOfSanctuary: true,
	blessingOfWisdom: TristateEffect.TristateEffectImproved,
	judgementsOfTheWise: true,
	vampiricTouch: true,
});

export const FULL_DEBUFFS = Debuffs.create({
	bloodFrenzy: true,
	curseOfElements: true,
	curseOfWeakness: TristateEffect.TristateEffectImproved,
	earthAndMoon: true,
	ebonPlaguebringer: true,
	exposeArmor: true,
	faerieFire: TristateEffect.TristateEffectImproved,
	giftOfArthas: true,
	heartOfTheCrusader: true,
	improvedScorch: true,
	insectSwarm: true,
	judgementOfLight: true,
	judgementOfWisdom: true,
	mangle: true,
	misery: true,
	scorpidSting: true,
	shadowEmbrace: true,
	shadowMastery: true,
	sunderArmor: true,
	thunderClap: TristateEffect.TristateEffectImproved,
	totemOfWrath: true,
	vindication: true,
});

// A copy of the spec's defaults in which every field is at least as strong as in the full set.
export function withFullRaid<T extends object>(type: IMessageType<T>, specDefaults: T, full: T): T {
	const merged = type.clone(specDefaults) as Record<string, unknown>;
	for (const [field, value] of Object.entries(full)) {
		if (typeof value == 'boolean') {
			merged[field] = merged[field] || value;
		} else if (typeof value == 'number') {
			merged[field] = Math.max(Number(merged[field]) || 0, value);
		}
	}
	return merged as T;
}
