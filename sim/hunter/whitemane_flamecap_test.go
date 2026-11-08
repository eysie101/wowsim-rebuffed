package hunter

import (
	"strings"
	"testing"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func TestWhitemaneFlameCapProcsOnRangedAutoShots(t *testing.T) {
	const own = "] [Hunter (#1)] "
	windows, shots, procs, specials, misplaced := 0, 0, 0, 0, 0
	for seed := int64(1); seed <= 3; seed++ {
		rsr := &proto.RaidSimRequest{
			Raid: core.SinglePlayerRaidProto(
				&proto.Player{
					Name:          "Hunter",
					Race:          proto.Race_RaceOrc,
					Class:         proto.Class_ClassHunter,
					Equipment:     core.GetGearSet("../../ui/hunter/gear_sets", "p1_mm").GearSet,
					Consumes:      FullConsumes,
					Spec:          PlayerOptionsBasic,
					Glyphs:        MMGlyphs,
					TalentsString: MMTalents,
					Buffs:         core.FullIndividualBuffs,
					Rotation:      core.GetAplRotation("../../ui/hunter/apls", "mm").Rotation,
				},
				core.FullPartyBuffs,
				core.FullRaidBuffs,
				core.FullDebuffs),
			Encounter:  &proto.Encounter{Duration: 300, Targets: []*proto.Target{core.NewDefaultTarget()}},
			SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: seed, Debug: true},
		}
		result := core.RunRaidSim(rsr)
		if result.ErrorResult != "" {
			t.Fatal(result.ErrorResult)
		}
		active, start, shot, proc, lastShot := false, "", 0, 0, ""
		for _, line := range strings.Split(result.Logs, "\n") {
			if !strings.Contains(line, own) || strings.Contains(line, "[DEBUG]") {
				continue
			}
			stamp := strings.SplitN(line, " ", 2)[0]
			landed := strings.Contains(line, " damage.") && !strings.Contains(line, "for 0.000 damage")
			switch {
			case strings.Contains(line, "Aura gained: {ItemID: 22788}"):
				active, start, shot, proc, lastShot = true, stamp, 0, 0, ""
				windows++
			case strings.Contains(line, "Aura faded: {ItemID: 22788}"):
				if active && proc != shot {
					t.Errorf("seed %d, Flame Cap %s to %s: %d Flamecap Fire procs for %d landed auto shots", seed, start, stamp, proc, shot)
				}
				active = false
			case !strings.Contains(line, own+"[Target"):
			case strings.Contains(line, "{ItemID: 22788}"):
				procs++
				proc++
				if stamp != lastShot {
					misplaced++
				}
			case strings.Contains(line, "{OtherID: 4}"):
				if active && landed {
					shot++
					shots++
					lastShot = stamp
				}
			case strings.Contains(line, "{SpellID: "):
				if active && landed {
					specials++
				}
			}
		}
	}
	if windows < 3 || shots < 100 || specials < 50 {
		t.Fatalf("parser not calibrated: %d Flame Cap windows, %d landed auto shots and %d landed specials in them", windows, shots, specials)
	}
	if misplaced > 0 {
		t.Errorf("%d of %d Flamecap Fire procs came at a moment without a landed auto shot", misplaced, procs)
	}
	t.Logf("%d Flame Cap windows: %d landed auto shots, %d Flamecap Fire procs, %d landed specials", windows, shots, procs, specials)
}

func TestWhitemaneFlameCapNotOnRaptorStrike(t *testing.T) {
	const own = "] [Hunter (#1)] "
	raptors, atRaptor, shots, procs := 0, 0, 0, 0
	for seed := int64(1); seed <= 3; seed++ {
		rotation := core.GetAplRotation("../../ui/hunter/apls", "mm").Rotation
		rotation.PriorityList = append([]*proto.APLListItem{{Action: &proto.APLAction{Action: &proto.APLAction_CastSpell{
			CastSpell: &proto.APLActionCastSpell{SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: 48996}}}}}}}, rotation.PriorityList...)
		rsr := &proto.RaidSimRequest{
			Raid: core.SinglePlayerRaidProto(
				&proto.Player{
					Name:          "Hunter",
					Race:          proto.Race_RaceOrc,
					Class:         proto.Class_ClassHunter,
					Equipment:     core.GetGearSet("../../ui/hunter/gear_sets", "p1_mm").GearSet,
					Consumes:      FullConsumes,
					Spec:          PlayerOptionsBasic,
					Glyphs:        MMGlyphs,
					TalentsString: MMTalents,
					Buffs:         core.FullIndividualBuffs,
					Rotation:      rotation,
				},
				core.FullPartyBuffs,
				core.FullRaidBuffs,
				core.FullDebuffs),
			Encounter:  &proto.Encounter{Duration: 300, Targets: []*proto.Target{core.NewDefaultTarget()}},
			SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: seed, Debug: true},
		}
		result := core.RunRaidSim(rsr)
		if result.ErrorResult != "" {
			t.Fatal(result.ErrorResult)
		}
		active, lastShot, lastRaptor := false, "", ""
		for _, line := range strings.Split(result.Logs, "\n") {
			if !strings.Contains(line, own) || strings.Contains(line, "[DEBUG]") {
				continue
			}
			stamp := strings.SplitN(line, " ", 2)[0]
			landed := strings.Contains(line, " damage.") && !strings.Contains(line, "for 0.000 damage")
			switch {
			case strings.Contains(line, "Aura gained: {ItemID: 22788}"):
				active = true
			case strings.Contains(line, "Aura faded: {ItemID: 22788}"):
				active = false
			case !active || !strings.Contains(line, own+"[Target"):
			case strings.Contains(line, "{ItemID: 22788}"):
				procs++
				if stamp == lastRaptor && stamp != lastShot {
					atRaptor++
				}
			case strings.Contains(line, "{OtherID: 4}") && landed:
				shots++
				lastShot = stamp
			case strings.Contains(line, "{SpellID: 48996}") && landed:
				raptors++
				lastRaptor = stamp
			}
		}
	}
	if raptors < 10 || shots < 100 || procs < shots/2 {
		t.Fatalf("parser not calibrated: %d landed Raptor Strikes, %d landed auto shots and %d Flamecap Fire procs in Flame Cap windows", raptors, shots, procs)
	}
	if atRaptor > 0 {
		t.Errorf("%d Flamecap Fire procs came with %d landed Raptor Strikes and no auto shot at the same moment", atRaptor, raptors)
	}
	t.Logf("Flame Cap windows: %d landed Raptor Strikes, %d landed auto shots, %d Flamecap Fire procs at a Raptor Strike", raptors, shots, atRaptor)
}
