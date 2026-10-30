package enhancement

import (
	"math"
	"strconv"
	"strings"
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
	const tolerance = 0.011
	stamp := func(line string) float64 {
		v, _ := strconv.ParseFloat(strings.Trim(strings.SplitN(line, " ", 2)[0], "[]"), 64)
		return v
	}
	var firsts, gaps []float64
	lives, novas, blasts, badBlasts := 0, 0, 0, 0
	casts := map[int32]float64{}
	shield := false
	for seed := int64(1); seed <= 10; seed++ {
		rsr := whitemaneRequest(1)
		rsr.SimOptions.RandomSeed = seed
		rsr.SimOptions.Debug = true
		result := core.RunRaidSim(rsr)
		if result.ErrorResult != "" {
			t.Fatal(result.ErrorResult)
		}
		summonAt, lastNova := -1.0, -1.0
		for _, line := range strings.Split(result.Logs, "\n") {
			if !strings.Contains(line, "] [Enha (#1) - Greater Fire Elemental] ") {
				continue
			}
			switch {
			case strings.Contains(line, "Pet summoned"):
				lives++
				summonAt, lastNova = stamp(line), -1
			case strings.Contains(line, "Casting {SpellID: 12470}"):
				novas++
				if lastNova < 0 {
					firsts = append(firsts, stamp(line)-summonAt)
				} else {
					gaps = append(gaps, stamp(line)-lastNova)
				}
				lastNova = stamp(line)
			case strings.Contains(line, "Casting {SpellID: 57984}"):
				blasts++
				if lastNova < 0 || math.Abs(stamp(line)-lastNova-2) > tolerance {
					badBlasts++
				}
			}
		}
		for _, p := range result.RaidMetrics.Parties[0].Players[0].Pets {
			if p.Name != "Greater Fire Elemental" {
				continue
			}
			for _, a := range p.Actions {
				shield = shield || a.Id.GetSpellId() == 13376
				for _, tgt := range a.Targets {
					casts[a.Id.GetSpellId()] += float64(tgt.Casts)
				}
			}
		}
	}
	if lives < 10 || len(gaps) < 100 {
		t.Fatalf("parser not calibrated: %d lives, %d Nova gaps in 10 fights", lives, len(gaps))
	}
	for _, d := range append(append([]float64{}, firsts...), gaps...) {
		if d < 8-tolerance || d > 10+tolerance {
			t.Errorf("Fire Nova %.2f s after the summon or the previous Nova, want 8 to 10 s", d)
		}
	}
	lo, hi, sum := math.Inf(1), math.Inf(-1), 0.0
	for _, d := range gaps {
		lo, hi, sum = math.Min(lo, d), math.Max(hi, d), sum+d
	}
	if mean := sum / float64(len(gaps)); lo > 8.3 || hi < 9.7 || mean < 8.7 || mean > 9.3 {
		t.Errorf("Nova gaps %.2f to %.2f s, mean %.3f s; want them spread over 8 to 10 s around 9 s", lo, hi, mean)
	}
	if badBlasts > 0 || blasts < novas-lives || blasts > novas {
		t.Errorf("%d Fire Blasts for %d Novas in %d lives, %d of them not right after a Nova's cast", blasts, novas, lives, badBlasts)
	}
	for _, spellID := range []int32{13339, 11350} {
		if _, found := casts[spellID]; found {
			t.Errorf("spell %d is a stock id; the Whitemane elemental does not cast it", spellID)
		}
	}
	if !shield {
		t.Error("no Fire Shield 13376 damage in the metrics")
	}
	t.Logf("%d lives: %.2f Novas and %.2f Blasts per life, Nova gaps %.2f to %.2f s", lives, float64(novas)/float64(lives), float64(blasts)/float64(lives), lo, hi)
}
