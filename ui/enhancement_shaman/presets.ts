import {
	Consumes,
	Flask,
	Food,
	Glyphs,
	IndividualBuffs,
	PartyBuffs,
	Potions,
	Profession,
	Race,
	RaidBuffs,
	Spec,
	TristateEffect,
	Debuffs,
} from '../core/proto/common.js';
import { SavedTalents } from '../core/proto/ui.js';

import {
	AirTotem,
	EarthTotem,
	FireTotem,
	WaterTotem,
	ShamanTotems,
	ShamanImbue,
	ShamanShield,
	ShamanSyncType,
	ShamanMajorGlyph,
	EnhancementShaman_Options as EnhancementShamanOptions,
} from '../core/proto/shaman.js';

import * as PresetUtils from '../core/preset_utils.js';
import { FULL_DEBUFFS, FULL_INDIVIDUAL_BUFFS, FULL_PARTY_BUFFS, FULL_RAID_BUFFS } from '../core/constants/full_raid_buffs.js';

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

import EnhaP1SpellGear from './gear_sets/enha_p1_spell.gear.json';
export const ENHA_P1_SPELL_PRESET = PresetUtils.makePresetGear('ENHA P1 Spell', EnhaP1SpellGear);
import EnhaP1WfGear from './gear_sets/enha_p1_wf.gear.json';
export const ENHA_P1_WF_PRESET = PresetUtils.makePresetGear('ENHA P1 WF', EnhaP1WfGear);

import EnhaP1SpellApl from './apls/enha_p1_spell.apl.json';
export const ROTATION_ENHA_P1_SPELL = PresetUtils.makePresetAPLRotation('ENHA P1 Spell', EnhaP1SpellApl);
import EnhaP1WfApl from './apls/enha_p1_wf.apl.json';
export const ROTATION_ENHA_P1_WF = PresetUtils.makePresetAPLRotation('ENHA P1 WF', EnhaP1WfApl);
import EnhaP1SpellAoe3Apl from './apls/enha_p1_spell_aoe3.apl.json';
export const ROTATION_ENHA_P1_SPELL_AOE3 = PresetUtils.makePresetAPLRotation('ENHA P1 Spell AoE 3+', EnhaP1SpellAoe3Apl);
import EnhaP1WfAoe3Apl from './apls/enha_p1_wf_aoe3.apl.json';
export const ROTATION_ENHA_P1_WF_AOE3 = PresetUtils.makePresetAPLRotation('ENHA P1 WF AoE 3+', EnhaP1WfAoe3Apl);

import WowsimsP1SpellApl from './apls/default_ft.apl.json';
export const ROTATION_WOWSIMS_P1_SPELL = PresetUtils.makePresetAPLRotation('wowsims P1 Spell', WowsimsP1SpellApl);
import WowsimsP1WfApl from './apls/default_wf.apl.json';
export const ROTATION_WOWSIMS_P1_WF = PresetUtils.makePresetAPLRotation('wowsims P1 WF', WowsimsP1WfApl);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/wotlk/talent-calc and copy the numbers in the url.
export const EnhaP1SpellTalents = {
	name: 'ENHA P1 Spell',
	data: SavedTalents.create({
		talentsString: '053030152-00505033105001333031130231051',
		glyphs: Glyphs.create({
			major1: ShamanMajorGlyph.GlyphOfFireNova,
			major2: ShamanMajorGlyph.GlyphOfFlametongueWeapon,
			major3: ShamanMajorGlyph.GlyphOfStormstrike,
			//minor glyphs dont affect damage done, all convenience/QoL
		})
	}),
};

export const EnhaP1WfTalents = {
	name: 'ENHA P1 WF',
	data: SavedTalents.create({
		talentsString: '053030152-00505033105001333031130231051',
		glyphs: Glyphs.create({
			major1: ShamanMajorGlyph.GlyphOfStormstrike,
			major2: ShamanMajorGlyph.GlyphOfFeralSpirit,
			major3: ShamanMajorGlyph.GlyphOfLightningShield,
			//minor glyphs dont affect damage done, all convenience/QoL
		})
	}),
};

export const DefaultOptions = EnhancementShamanOptions.create({
	shield: ShamanShield.LightningShield,
	imbueMh: ShamanImbue.FlametongueWeapon,
	imbueOh: ShamanImbue.FlametongueWeapon,
	syncType: ShamanSyncType.Auto,
	totems: ShamanTotems.create({
		earth: 	EarthTotem.StrengthOfEarthTotem,
		fire: 	FireTotem.MagmaTotem,
		water: 	WaterTotem.ManaSpringTotem,
		air: 	AirTotem.WindfuryTotem,
	})
});

export const DefaultConsumes = Consumes.create({
	defaultPotion: Potions.PotionOfSpeed,
	flask: Flask.FlaskOfTheFrostWyrm,
	food: Food.FoodFishFeast,
});

export const DefaultRaidBuffs = RaidBuffs.create({
	giftOfTheWild: TristateEffect.TristateEffectImproved,
	arcaneBrilliance: true,
	leaderOfThePack: TristateEffect.TristateEffectImproved,
	totemOfWrath: true,
	wrathOfAirTotem: true,
	moonkinAura: TristateEffect.TristateEffectImproved,
	sanctifiedRetribution: true,
	divineSpirit: true,
	battleShout: TristateEffect.TristateEffectImproved,
	demonicPactSp: 500,
});

export const DefaultDebuffs = Debuffs.create({
	bloodFrenzy: true,
	sunderArmor: true,
	curseOfWeakness: TristateEffect.TristateEffectRegular,
	curseOfElements: true,
	faerieFire: TristateEffect.TristateEffectImproved,
	judgementOfWisdom: true,
	misery: true,
	totemOfWrath: true,
	shadowMastery: true,
});

const enhaTotems = ShamanTotems.create({
	earth: EarthTotem.StrengthOfEarthTotem,
	air: AirTotem.WindfuryTotem,
	water: WaterTotem.HealingStreamTotem,
	fire: FireTotem.MagmaTotem,
	useFireElemental: true,
});
const enhaSpellOptions = EnhancementShamanOptions.create({
	shield: ShamanShield.LightningShield,
	syncType: ShamanSyncType.Auto,
	imbueMh: ShamanImbue.FlametongueWeapon,
	imbueOh: ShamanImbue.FlametongueWeapon,
	totems: enhaTotems,
});
const enhaWfOptions = EnhancementShamanOptions.create({ ...enhaSpellOptions, imbueMh: ShamanImbue.WindfuryWeapon });
const enhaSpellConsumes = Consumes.create({
	flask: Flask.FlaskOfTheFrostWyrm,
	food: Food.FoodFishFeast,
	defaultPotion: Potions.PotionOfSpeed,
	prepopPotion: Potions.PotionOfSpeed,
});
const enhaWfConsumes = Consumes.create({ ...enhaSpellConsumes, flask: Flask.FlaskOfEndlessRage });
const enhaRaid = {
	race: Race.RaceOrc,
	professions: [Profession.Engineering, Profession.Tailoring],
	raidBuffs: RaidBuffs.create({ ...FULL_RAID_BUFFS, manaSpringTotem: TristateEffect.TristateEffectMissing }),
	partyBuffs: PartyBuffs.create({ ...FULL_PARTY_BUFFS, manaTideTotems: 0, heroicPresence: true }),
	individualBuffs: IndividualBuffs.create({ ...FULL_INDIVIDUAL_BUFFS, judgementsOfTheWise: false }),
	debuffs: Debuffs.create({ ...FULL_DEBUFFS }),
};
const enhaWfAoeTalents = SavedTalents.create({
	...EnhaP1WfTalents.data,
	glyphs: Glyphs.create({
		major1: ShamanMajorGlyph.GlyphOfFireNova,
		major2: ShamanMajorGlyph.GlyphOfFlametongueWeapon,
		major3: ShamanMajorGlyph.GlyphOfFeralSpirit,
	}),
});

export const BUILD_ENHA_P1_SPELL: PresetUtils.PresetBuild<Spec.SpecEnhancementShaman> = {
	name: 'ENHA P1 Spell', gear: ENHA_P1_SPELL_PRESET, talents: EnhaP1SpellTalents.data, rotation: ROTATION_ENHA_P1_SPELL,
	specOptions: enhaSpellOptions, consumes: enhaSpellConsumes, ...enhaRaid,
	encounter: { duration: 240, durationVariation: 5, numTargets: 1 },
};
export const BUILD_ENHA_P1_WF: PresetUtils.PresetBuild<Spec.SpecEnhancementShaman> = {
	name: 'ENHA P1 WF', gear: ENHA_P1_WF_PRESET, talents: EnhaP1WfTalents.data, rotation: ROTATION_ENHA_P1_WF,
	specOptions: enhaWfOptions, consumes: enhaWfConsumes, ...enhaRaid,
	encounter: { duration: 240, durationVariation: 5, numTargets: 1 },
};
export const BUILD_ENHA_P1_SPELL_AOE3: PresetUtils.PresetBuild<Spec.SpecEnhancementShaman> = {
	name: 'ENHA P1 Spell AoE 3+', gear: ENHA_P1_SPELL_PRESET, talents: EnhaP1SpellTalents.data, rotation: ROTATION_ENHA_P1_SPELL_AOE3,
	specOptions: enhaSpellOptions, consumes: enhaSpellConsumes, ...enhaRaid,
	encounter: { duration: 240, durationVariation: 5, numTargets: 3 },
};
export const BUILD_ENHA_P1_WF_AOE3: PresetUtils.PresetBuild<Spec.SpecEnhancementShaman> = {
	name: 'ENHA P1 WF AoE 3+', gear: ENHA_P1_WF_PRESET, talents: enhaWfAoeTalents, rotation: ROTATION_ENHA_P1_WF_AOE3,
	specOptions: enhaWfOptions, consumes: enhaWfConsumes, ...enhaRaid,
	encounter: { duration: 240, durationVariation: 5, numTargets: 3 },
};

import WowsimsP1Gear from './gear_sets/p1.gear.json';
export const WOWSIMS_P1_GEAR = PresetUtils.makePresetGear('wowsims P1', WowsimsP1Gear);
export const WowsimsP1Talents = {
	name: 'wowsims P1',
	data: SavedTalents.create({
		talentsString: '053030152-30405003105021333031131031051',
		glyphs: Glyphs.create({
			major1: ShamanMajorGlyph.GlyphOfFireNova,
			major2: ShamanMajorGlyph.GlyphOfFlametongueWeapon,
			major3: ShamanMajorGlyph.GlyphOfFeralSpirit,
		}),
	}),
};
const wowsimsTotems = ShamanTotems.create({
	earth: EarthTotem.StrengthOfEarthTotem,
	fire: FireTotem.MagmaTotem,
	water: WaterTotem.ManaSpringTotem,
	air: AirTotem.WindfuryTotem,
});
const wowsimsConsumes = Consumes.create({
	defaultPotion: Potions.PotionOfSpeed,
	flask: Flask.FlaskOfEndlessRage,
	food: Food.FoodFishFeast,
});

export const BUILD_WOWSIMS_P1_SPELL: PresetUtils.PresetBuild<Spec.SpecEnhancementShaman> = {
	name: 'wowsims P1 Spell', gear: WOWSIMS_P1_GEAR, talents: WowsimsP1Talents.data, rotation: ROTATION_WOWSIMS_P1_SPELL,
	specOptions: EnhancementShamanOptions.create({
		shield: ShamanShield.LightningShield,
		syncType: ShamanSyncType.Auto,
		imbueMh: ShamanImbue.FlametongueWeaponDownrank,
		imbueOh: ShamanImbue.FlametongueWeapon,
		totems: ShamanTotems.create({ ...wowsimsTotems, useFireElemental: true }),
	}),
	consumes: wowsimsConsumes,
};
export const BUILD_WOWSIMS_P1_WF: PresetUtils.PresetBuild<Spec.SpecEnhancementShaman> = {
	name: 'wowsims P1 WF', gear: WOWSIMS_P1_GEAR, talents: WowsimsP1Talents.data, rotation: ROTATION_WOWSIMS_P1_WF,
	specOptions: EnhancementShamanOptions.create({
		shield: ShamanShield.LightningShield,
		syncType: ShamanSyncType.Auto,
		imbueMh: ShamanImbue.WindfuryWeapon,
		imbueOh: ShamanImbue.FlametongueWeapon,
		totems: wowsimsTotems,
	}),
	consumes: wowsimsConsumes,
};
