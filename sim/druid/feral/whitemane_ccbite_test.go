package feral

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func TestWhitemaneClearcastingBite(t *testing.T) {
	spent := regexp.MustCompile(`Spent ([\d.]+) energy from \{SpellID: 48577\}`)
	free, paid, badFree, badPaid := 0, 0, 0, 0
	for seed := int64(11); seed <= 20; seed++ {
		rsr := &proto.RaidSimRequest{
			Raid: core.SinglePlayerRaidProto(
				&proto.Player{
					Name:      "Feral",
					Race:      proto.Race_RaceTauren,
					Class:     proto.Class_ClassDruid,
					Equipment: core.GetGearSet("../../../ui/feral_druid/gear_sets", "p1_feral").GearSet,
					Consumes:  FullConsumes,
					Spec:      PlayerOptionsMonoCatNoBleed,
					Buffs:     core.FullIndividualBuffs,
					Glyphs:    StandardGlyphs,
					Rotation:  core.GetAplRotation("../../../ui/feral_druid/apls", "meow_feral").Rotation,

					TalentsString:   StandardTalents,
					InFrontOfTarget: true,
				},
				core.FullPartyBuffs,
				core.FullRaidBuffs,
				core.FullDebuffs),
			Encounter:  &proto.Encounter{Duration: 240, Targets: []*proto.Target{core.NewDefaultTarget()}},
			SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: seed, Debug: true},
		}
		result := core.RunRaidSim(rsr)
		if result.ErrorResult != "" {
			t.Fatal(result.ErrorResult)
		}
		groups, order := map[string][]string{}, []string{}
		for _, line := range strings.Split(result.Logs, "\n") {
			if !strings.Contains(line, "] [Feral (#1)] ") {
				continue
			}
			stamp := strings.SplitN(line, " ", 2)[0]
			if _, seen := groups[stamp]; !seen {
				order = append(order, stamp)
			}
			groups[stamp] = append(groups[stamp], line)
		}
		for _, stamp := range order {
			cast, isFree, energy := false, false, 0.0
			for _, l := range groups[stamp] {
				if strings.Contains(l, "Casting {SpellID: 48577}") {
					cast, isFree = true, strings.Contains(l, "(Cost = 0.000,")
				}
				if m := spent.FindStringSubmatch(l); m != nil {
					v, _ := strconv.ParseFloat(m[1], 64)
					energy += v
				}
			}
			switch {
			case !cast:
			case isFree:
				free++
				if energy != 0 {
					badFree++
					if badFree <= 3 {
						t.Errorf("seed %d, instant %s: free Bite spent %.3f energy:\n%s", seed, stamp, energy, strings.Join(groups[stamp], "\n"))
					}
				}
			default:
				paid++
				if energy == 0 {
					badPaid++
				}
			}
		}
	}
	if paid < 20 || badPaid > 0 {
		t.Fatalf("parser not calibrated: %d paid Bites, %d of them without an energy spend", paid, badPaid)
	}
	if free < 20 {
		t.Fatalf("only %d free (Clearcasting) Bites in 10 fights of 240 s; the rotation must take a due Bite on Clearcasting", free)
	}
	if badFree > 0 {
		t.Fatalf("%d of %d free Bites spent energy", badFree, free)
	}
	t.Logf("%d free Bites without an energy spend, %d paid Bites with one", free, paid)
}
