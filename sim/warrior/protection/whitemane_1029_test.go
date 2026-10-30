package protection

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

func TestWhitemaneGlyphOfSunderArmorDevastate(t *testing.T) {
	stacks := regexp.MustCompile(`\] \[Target 1\] \{SpellID: 47467\} stacks: \d+ --> (\d+)`)
	debuffs := googleProto.Clone(core.FullDebuffs).(*proto.Debuffs)
	debuffs.SunderArmor, debuffs.ExposeArmor = false, false
	devastates, atFive, applied, early := 0, 0, 0, 0
	for seed := int64(1); seed <= 3; seed++ {
		rsr := &proto.RaidSimRequest{
			Raid: core.SinglePlayerRaidProto(
				&proto.Player{
					Name:          "Prot",
					Race:          proto.Race_RaceOrc,
					Class:         proto.Class_ClassWarrior,
					Equipment:     core.GetGearSet("../../../ui/protection_warrior/gear_sets", "p1_balanced").GearSet,
					Consumes:      FullConsumes,
					Spec:          PlayerOptionsBasic,
					Buffs:         core.FullIndividualBuffs,
					TalentsString: DefaultTalents,
					Glyphs: &proto.Glyphs{
						Major1: int32(proto.WarriorMajorGlyph_GlyphOfSunderArmor),
						Major2: int32(proto.WarriorMajorGlyph_GlyphOfDevastate),
						Major3: int32(proto.WarriorMajorGlyph_GlyphOfVigilance),
					},
					Rotation:        core.GetAplRotation("../../../ui/protection_warrior/apls", "default").Rotation,
					InFrontOfTarget: true,
				},
				core.FullPartyBuffs,
				core.FullRaidBuffs,
				debuffs),
			Encounter:  &proto.Encounter{Duration: 120, Targets: []*proto.Target{core.NewDefaultTarget(), core.NewDefaultTarget()}},
			SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: seed, Debug: true},
		}
		result := core.RunRaidSim(rsr)
		if result.ErrorResult != "" {
			t.Fatal(result.ErrorResult)
		}
		first, firstAtDevastate, lastApplied := 0, 0, -1
		for _, line := range strings.Split(result.Logs, "\n") {
			switch {
			case strings.Contains(line, "] [Prot (#1)] [Target 1] {SpellID: 47498} ") && strings.Contains(line, " damage. (Threat"):
				devastates++
				firstAtDevastate = first
				if first == 5 {
					atFive++
				}
			case stacks.MatchString(line):
				first, _ = strconv.Atoi(stacks.FindStringSubmatch(line)[1])
			case strings.Contains(line, "] [Target 2] ") && strings.Contains(line, "{SpellID: 47467}") && lastApplied != devastates:
				lastApplied = devastates
				applied++
				if firstAtDevastate != 5 {
					early++
				}
			}
		}
	}
	if devastates < 50 || atFive < 30 {
		t.Fatalf("parser not calibrated: %d landed Devastates on the first target, %d of them at 5 Sunder stacks", devastates, atFive)
	}
	if applied != atFive || early > 0 {
		t.Fatalf("%d landed Devastates on a first target with 5 Sunder stacks, %d Sunder applications on the second target, %d of them from a Devastate below 5 stacks", atFive, applied, early)
	}
	t.Logf("%d landed Devastates, %d at 5 stacks, each applied Sunder Armor to the second target", devastates, atFive)
}
