import { GemColor, HandType, ItemType, RangedWeaponType, Stat, WeaponType } from '../proto/common.js';
import { UIItem as Item } from '../proto/ui.js';

import { armorTypeNames, classNames, rangedWeaponTypeNames, weaponTypeNames } from './names.js';

const OWN_TOOLTIP_IDS = new Set([130031, 131001, 131002, 132001, 132003]);

const WHITEMANE_STATS_IDS = new Set([42569]);

const PRIMARY: Array<[Stat, string]> = [
	[Stat.StatStrength, 'Strength'],
	[Stat.StatAgility, 'Agility'],
	[Stat.StatStamina, 'Stamina'],
	[Stat.StatIntellect, 'Intellect'],
	[Stat.StatSpirit, 'Spirit'],
];

const RESISTANCES: Array<[Stat, string]> = [
	[Stat.StatArcaneResistance, 'Arcane Resistance'],
	[Stat.StatFireResistance, 'Fire Resistance'],
	[Stat.StatFrostResistance, 'Frost Resistance'],
	[Stat.StatNatureResistance, 'Nature Resistance'],
	[Stat.StatShadowResistance, 'Shadow Resistance'],
];

const EQUIP: Array<[Stat, (n: number) => string]> = [
	[Stat.StatDefense, n => `Increases defense rating by ${n}.`],
	[Stat.StatDodge, n => `Increases your dodge rating by ${n}.`],
	[Stat.StatParry, n => `Increases your parry rating by ${n}.`],
	[Stat.StatBlock, n => `Increases your shield block rating by ${n}.`],
	[Stat.StatBlockValue, n => `Increases the block value of your shield by ${n}.`],
	[Stat.StatMeleeHit, n => `Increases your hit rating by ${n}.`],
	[Stat.StatMeleeCrit, n => `Increases your critical strike rating by ${n}.`],
	[Stat.StatMeleeHaste, n => `Increases your haste rating by ${n}.`],
	[Stat.StatExpertise, n => `Increases your expertise rating by ${n}.`],
	[Stat.StatResilience, n => `Improves your resilience rating by ${n}.`],
	[Stat.StatAttackPower, n => `Increases attack power by ${n}.`],
	[Stat.StatArmorPenetration, n => `Increases your armor penetration rating by ${n}.`],
	[Stat.StatSpellPower, n => `Increases spell power by ${n}.`],
	[Stat.StatSpellPenetration, n => `Increases your spell penetration by ${n}.`],
	[Stat.StatMP5, n => `Restores ${n} mana per 5 sec.`],
];

const SOCKET_BONUS: Array<[Stat, string]> = [
	...PRIMARY,
	[Stat.StatSpellPower, 'Spell Power'],
	[Stat.StatMP5, 'Mana every 5 seconds'],
	[Stat.StatMeleeHit, 'Hit Rating'],
	[Stat.StatMeleeCrit, 'Critical Strike Rating'],
	[Stat.StatMeleeHaste, 'Haste Rating'],
	[Stat.StatAttackPower, 'Attack Power'],
	[Stat.StatArmorPenetration, 'Armor Penetration Rating'],
	[Stat.StatExpertise, 'Expertise Rating'],
	[Stat.StatDefense, 'Defense Rating'],
	[Stat.StatDodge, 'Dodge Rating'],
	[Stat.StatParry, 'Parry Rating'],
	[Stat.StatBlock, 'Block Rating'],
	[Stat.StatBlockValue, 'Block Value'],
	[Stat.StatResilience, 'Resilience Rating'],
];

const SLOTS: Partial<Record<ItemType, string>> = {
	[ItemType.ItemTypeHead]: 'Head',
	[ItemType.ItemTypeNeck]: 'Neck',
	[ItemType.ItemTypeShoulder]: 'Shoulder',
	[ItemType.ItemTypeBack]: 'Back',
	[ItemType.ItemTypeChest]: 'Chest',
	[ItemType.ItemTypeWrist]: 'Wrist',
	[ItemType.ItemTypeHands]: 'Hands',
	[ItemType.ItemTypeWaist]: 'Waist',
	[ItemType.ItemTypeLegs]: 'Legs',
	[ItemType.ItemTypeFeet]: 'Feet',
	[ItemType.ItemTypeFinger]: 'Finger',
	[ItemType.ItemTypeTrinket]: 'Trinket',
};

const HANDS: Partial<Record<HandType, string>> = {
	[HandType.HandTypeMainHand]: 'Main Hand',
	[HandType.HandTypeOneHand]: 'One-Hand',
	[HandType.HandTypeOffHand]: 'Off Hand',
	[HandType.HandTypeTwoHand]: 'Two-Hand',
};

const SOCKETS: Partial<Record<GemColor, string>> = {
	[GemColor.GemColorMeta]: 'Meta Socket',
	[GemColor.GemColorRed]: 'Red Socket',
	[GemColor.GemColorBlue]: 'Blue Socket',
	[GemColor.GemColorYellow]: 'Yellow Socket',
	[GemColor.GemColorPrismatic]: 'Prismatic Socket',
};

const RELICS = new Set([RangedWeaponType.RangedWeaponTypeIdol, RangedWeaponType.RangedWeaponTypeLibram, RangedWeaponType.RangedWeaponTypeTotem, RangedWeaponType.RangedWeaponTypeSigil]);

function escapeHtml(s: string): string {
	return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function slotAndType(item: Item): [string, string] {
	if (item.type == ItemType.ItemTypeWeapon) {
		if (item.weaponType == WeaponType.WeaponTypeShield) return ['Off Hand', 'Shield'];
		if (item.weaponType == WeaponType.WeaponTypeOffHand) return ['Held In Off-hand', ''];
		return [HANDS[item.handType] || '', weaponTypeNames.get(item.weaponType) || ''];
	}
	if (item.type == ItemType.ItemTypeRanged) {
		if (RELICS.has(item.rangedWeaponType)) return ['Relic', rangedWeaponTypeNames.get(item.rangedWeaponType) || ''];
		if (item.rangedWeaponType == RangedWeaponType.RangedWeaponTypeThrown) return ['Thrown', ''];
		return ['Ranged', rangedWeaponTypeNames.get(item.rangedWeaponType) || ''];
	}
	return [SLOTS[item.type] || '', item.type == ItemType.ItemTypeBack || !item.armorType ? '' : armorTypeNames.get(item.armorType) || ''];
}

export function buildItemTooltip(item: Item, lookup: (id: number) => Item | null): string {
	const stat = (s: Stat) => Math.round(item.stats[s] || 0);
	const lines: Array<string> = [];
	lines.push(`<b class="q${item.quality}">${escapeHtml(item.name)}</b>`);
	if (item.heroic) lines.push('<span class="q2">Heroic</span>');
	if (item.upgradeLevel) lines.push(`<span class="q2">Upgrade level +${item.upgradeLevel}</span>`);
	if (item.epicVariant) lines.push('<span class="q2">Epic version</span>');
	lines.push(`<span class="q">Item Level ${item.ilvl}</span>`);
	if (item.unique) lines.push('Unique');

	const [slot, type] = slotAndType(item);
	lines.push(`<table width="100%"><tr><td>${slot}</td><th>${type}</th></tr></table>`);

	if (item.weaponSpeed > 0) {
		const dps = (item.weaponDamageMin + item.weaponDamageMax) / 2 / item.weaponSpeed;
		lines.push(
			`<table width="100%"><tr><td>${Math.round(item.weaponDamageMin)} - ${Math.round(item.weaponDamageMax)} Damage</td>` +
				`<th>Speed ${item.weaponSpeed.toFixed(2)}</th></tr></table>`,
		);
		lines.push(`(${dps.toFixed(1)} damage per second)`);
	}
	const armor = stat(Stat.StatArmor) + stat(Stat.StatBonusArmor);
	if (armor > 0) lines.push(stat(Stat.StatBonusArmor) > 0 ? `<span class="q2">${armor} Armor</span>` : `${armor} Armor`);
	for (const [s, name] of PRIMARY) {
		if (stat(s) > 0) lines.push(`+${stat(s)} ${name}`);
	}
	for (const [s, name] of RESISTANCES) {
		if (stat(s) > 0) lines.push(`+${stat(s)} ${name}`);
	}

	for (const color of item.gemSockets) {
		lines.push(`<span class="q0">${SOCKETS[color] || 'Socket'}</span>`);
	}
	const bonus = SOCKET_BONUS.filter(([s]) => (item.socketBonus[s] || 0) > 0).map(([s, name]) => `+${Math.round(item.socketBonus[s])} ${name}`);
	if (bonus.length) lines.push(`<span class="q0">Socket Bonus: ${bonus.join(', ')}</span>`);

	if (item.classAllowlist.length) {
		lines.push(`Classes: ${item.classAllowlist.map(c => classNames.get(c) || '').join(', ')}`);
	}

	for (const [s, text] of EQUIP) {
		if (stat(s) > 0) lines.push(`<span class="q2">Equip: ${text(stat(s))}</span>`);
	}
	const extraRangedAP = stat(Stat.StatRangedAttackPower) - stat(Stat.StatAttackPower);
	if (extraRangedAP > 0) lines.push(`<span class="q2">Equip: Increases ranged attack power by ${extraRangedAP}.</span>`);
	for (const text of item.effectTexts) lines.push(`<span class="q2">${escapeHtml(text).replace(/\n/g, '<br />')}</span>`);

	if (item.effectItemId) {
		const source = lookup(item.effectItemId);
		const scale = item.effectScale && item.effectScale != 1 ? `, ${Math.round((item.effectScale - 1) * 100)}% stronger` : '';
		if (source && (source.name != item.name || scale)) lines.push(`<span class="q2">Effect of ${escapeHtml(source.name)}${scale}</span>`);
	}
	if (item.setName) lines.push(`<br /><span class="q">${escapeHtml(item.setName)}</span>`);

	return `<table><tr><td>${lines.join('<br />')}</td></tr></table>`;
}

let itemLookup: (id: number) => Item | null = () => null;

export function wowheadBaseId(id: number): number {
	let item = itemLookup(id);
	for (let step = 0; item && item.id >= 100000 && item.variantBaseId && step < 3; step++) item = itemLookup(item.variantBaseId);
	return item && item.id < 100000 ? item.id : 0;
}

export function registerWhitemaneTooltips(lookup: (id: number) => Item | null) {
	itemLookup = lookup;
	(window as any).wsrItemTooltip = (id: number) => {
		if ((id < 100000 && !WHITEMANE_STATS_IDS.has(id)) || OWN_TOOLTIP_IDS.has(id)) return null;
		const item = lookup(id);
		if (!item) return null;
		return { html: buildItemTooltip(item, lookup), icon: item.icon };
	};
}
