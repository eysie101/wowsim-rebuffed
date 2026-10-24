package enhancement

import (
	"testing"
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func whitemaneRequest(iterations int32) *proto.RaidSimRequest {
	return &proto.RaidSimRequest{
		Raid: core.SinglePlayerRaidProto(
			&proto.Player{
				Name:          "Enha",
				Race:          proto.Race_RaceOrc,
				Class:         proto.Class_ClassShaman,
				Equipment:     core.GetGearSet("../../../ui/enhancement_shaman/gear_sets", "enha_p1_spell").GearSet,
				TalentsString: StandardTalents,
				Glyphs:        StandardGlyphs,
				Consumes:      FullConsumes,
				Spec:          PlayerOptionsFTFT,
				Buffs:         core.FullIndividualBuffs,
				Rotation:      core.GetAplRotation("../../../ui/enhancement_shaman/apls", "enha_p1_spell").Rotation,
			},
			core.FullPartyBuffs,
			core.FullRaidBuffs,
			core.FullDebuffs),
		Encounter: &proto.Encounter{
			Duration: 240,
			Targets:  []*proto.Target{core.NewDefaultTarget()},
		},
		SimOptions: &proto.SimOptions{Iterations: iterations, RandomSeed: 1},
	}
}

func TestWhitemaneFlurryDuration(t *testing.T) {
	rsr := whitemaneRequest(1)
	env, _, _ := core.NewEnvironment(rsr.Raid, rsr.Encounter, false)
	enh, ok := env.Raid.Parties[0].Players[0].(*EnhancementShaman)
	if !ok {
		t.Fatalf("player is %T, not an Enhancement shaman", env.Raid.Parties[0].Players[0])
	}
	aura := enh.GetAura("Flurry Proc")
	if aura == nil {
		t.Fatal("no Flurry Proc aura registered")
	}
	if aura.Duration != 15*time.Second {
		t.Fatalf("Flurry Proc duration %s, want 15s", aura.Duration)
	}
}

func TestWhitemaneFireElementalCadence(t *testing.T) {
	const iterations = 200
	result := core.RunRaidSim(whitemaneRequest(iterations))
	if result.ErrorResult != "" {
		t.Fatal(result.ErrorResult)
	}
	var pet *proto.UnitMetrics
	for _, p := range result.RaidMetrics.Parties[0].Players[0].Pets {
		if p.Name == "Greater Fire Elemental" {
			pet = p
		}
	}
	if pet == nil {
		t.Fatal("no Greater Fire Elemental in the metrics")
	}
	casts := map[int32]float64{}
	for _, a := range pet.Actions {
		for _, tgt := range a.Targets {
			casts[a.Id.GetSpellId()] += float64(tgt.Casts)
		}
	}
	for _, spellID := range []int32{57984, 12470} {
		if got := casts[spellID] / iterations; got != 16 {
			t.Errorf("spell %d: %.3f casts per life, want 16", spellID, got)
		}
	}
	for _, spellID := range []int32{13339, 11350} {
		if _, found := casts[spellID]; found {
			t.Errorf("spell %d is a stock id; the Whitemane elemental does not cast it", spellID)
		}
	}
	shield := false
	for _, a := range pet.Actions {
		shield = shield || a.Id.GetSpellId() == 13376
	}
	if !shield {
		t.Error("no Fire Shield 13376 damage in the metrics")
	}
}
