import {
	Consumes,
	Flask,
	Food,
	Glyphs,
	Potions,
	RaidBuffs,
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

// Preset options for this spec.
// Eventually we will import these values for the raid sim too, so its good to
// keep them in a separate file.

// Whitemane phase 1: the best-in-slot sets of the spell and Windfury builds (Jewelcrafting and Engineering)
import EnhaP1SpellGear from './gear_sets/enha_p1_spell.gear.json';
export const ENHA_P1_SPELL_PRESET = PresetUtils.makePresetGear('ENHA P1 Spell', EnhaP1SpellGear);
import EnhaP1WfGear from './gear_sets/enha_p1_wf.gear.json';
export const ENHA_P1_WF_PRESET = PresetUtils.makePresetGear('ENHA P1 WF', EnhaP1WfGear);

// The priority lists of the phase 1 search: single target and three or more targets, per build
import EnhaP1SpellApl from './apls/enha_p1_spell.apl.json';
export const ROTATION_ENHA_P1_SPELL = PresetUtils.makePresetAPLRotation('ENHA P1 Spell', EnhaP1SpellApl);
import EnhaP1WfApl from './apls/enha_p1_wf.apl.json';
export const ROTATION_ENHA_P1_WF = PresetUtils.makePresetAPLRotation('ENHA P1 WF', EnhaP1WfApl);
import EnhaP1SpellAoe3Apl from './apls/enha_p1_spell_aoe3.apl.json';
export const ROTATION_ENHA_P1_SPELL_AOE3 = PresetUtils.makePresetAPLRotation('ENHA P1 Spell AoE 3+', EnhaP1SpellAoe3Apl);
import EnhaP1WfAoe3Apl from './apls/enha_p1_wf_aoe3.apl.json';
export const ROTATION_ENHA_P1_WF_AOE3 = PresetUtils.makePresetAPLRotation('ENHA P1 WF AoE 3+', EnhaP1WfAoe3Apl);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wowhead.com/wotlk/talent-calc and copy the numbers in the url.
// Whitemane: Ancestral Knowledge 5/5 and Improved Stormstrike 2/2 in both builds; glyphs per build.
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
