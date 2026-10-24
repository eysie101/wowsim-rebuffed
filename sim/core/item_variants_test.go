package core

import (
	"testing"

	"github.com/wowsims/wotlk/sim/core/proto"
	"github.com/wowsims/wotlk/sim/core/stats"
)

func TestItemFromProtoSharedEffect(t *testing.T) {
	item := ItemFromProto(&proto.SimItem{Id: 1206830, EffectItemId: 37390, EffectScale: 1.1464})
	if item.ID != 37390 || item.VariantID != 1206830 || item.RealID() != 1206830 {
		t.Fatalf("ID %d, VariantID %d, RealID %d; want 37390, 1206830, 1206830", item.ID, item.VariantID, item.RealID())
	}
	if spec := item.ToItemSpecProto(); spec.Id != 1206830 {
		t.Fatalf("item spec ID %d, want 1206830", spec.Id)
	}

	plain := ItemFromProto(&proto.SimItem{Id: 37390})
	if plain.ID != 37390 || plain.VariantID != 0 || plain.RealID() != 37390 {
		t.Fatalf("plain item: ID %d, VariantID %d, RealID %d", plain.ID, plain.VariantID, plain.RealID())
	}
}

func TestScaleItemEffect(t *testing.T) {
	character := &Character{}
	bonus := stats.Stats{stats.MeleeHaste: 444, stats.SpellHaste: 444}

	if got := character.ScaleItemEffectStats(bonus); got[stats.MeleeHaste] != 444 {
		t.Fatalf("outside an item effect: %v, want 444", got[stats.MeleeHaste])
	}

	character.itemEffectScale = 1.1464
	got := character.ScaleItemEffectStats(bonus)
	if got[stats.MeleeHaste] != 509 || got[stats.SpellHaste] != 509 {
		t.Fatalf("scaled: %v / %v, want 509", got[stats.MeleeHaste], got[stats.SpellHaste])
	}
	if bonus[stats.MeleeHaste] != 444 {
		t.Fatalf("the original bonus changed to %v", bonus[stats.MeleeHaste])
	}
	if got := character.ScaleItemEffect(1024); got != 1174 {
		t.Fatalf("scaled damage %v, want 1174", got)
	}
}
