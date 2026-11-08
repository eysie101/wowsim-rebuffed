package sim

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

var logLineTime = regexp.MustCompile(`^\[(-?\d+\.\d+)\] `)

var bossCastLine = regexp.MustCompile(`^\[(-?\d+\.\d+)\] \[Target 1\] (?:Casting (\{[^}]*\}) \(Cost = [^,]*, Cast Time = ([^,]*),|Completed cast (\{[^}]*\}))`)

var bossSwingIDs = []string{"{OtherID: 3, ", "{SpellID: 63511, Tag: 1}", "{SpellID: 62867, Tag: 1}"}

type bossCastCount struct {
	iterations, hardcasts, completed, late, ended, cutOff, running, swings, inside int
}

func countBossCasts(t *testing.T, logs string) bossCastCount {
	type cast struct {
		id         string
		start, end float64
	}
	var c bossCastCount
	var pending *cast
	instant, previous := "", 1e9
	unfinished := func(at float64) {
		switch {
		case pending == nil:
		case at >= pending.end-0.011:
			c.ended++
		default:
			c.cutOff++
		}
	}
	for _, line := range strings.Split(logs, "\n") {
		tm := logLineTime.FindStringSubmatch(line)
		if tm == nil {
			continue
		}
		at, err := strconv.ParseFloat(tm[1], 64)
		if err != nil {
			t.Fatal(err)
		}
		if at < previous-1 {
			if pending != nil && previous < pending.end-0.011 {
				c.running++
			} else {
				unfinished(previous)
			}
			c.iterations++
			pending, instant = nil, ""
		}
		previous = at
		m := bossCastLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		if m[2] != "" {
			d, err := time.ParseDuration(m[3])
			if err != nil {
				t.Fatal(err)
			}
			if d == 0 {
				instant = m[2]
				for _, id := range bossSwingIDs {
					if strings.HasPrefix(m[2], id) {
						c.swings++
						if pending != nil && at > pending.start+0.011 && at < pending.end-0.011 {
							c.inside++
						}
						break
					}
				}
				continue
			}
			unfinished(at)
			c.hardcasts++
			pending = &cast{m[2], at, at + d.Seconds()}
			continue
		}
		if m[4] == instant {
			instant = ""
			continue
		}
		if pending != nil && m[4] == pending.id {
			c.completed++
			if at > pending.end+0.015 {
				c.late++
			}
			pending = nil
		}
	}
	if pending != nil && previous < pending.end-0.011 {
		c.running++
	} else {
		unfinished(previous)
	}
	return c
}

func bossCastRequest(t *testing.T, path string) *proto.RaidSimRequest {
	preset := core.GetPresetTargetWithPath(path)
	if preset == nil {
		t.Fatalf("no preset target %s", path)
	}
	raid := core.SinglePlayerRaidProto(&proto.Player{
		Name:          "Tank",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassWarrior,
		Equipment:     core.GetGearSet("../ui/protection_warrior/gear_sets", "p1_balanced").GearSet,
		TalentsString: "2500030023-302-053351225000012521030113321",
		Spec: &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{
			Options: &proto.ProtectionWarrior_Options{Shout: proto.WarriorShout_WarriorShoutCommanding},
		}},
		Buffs:           core.FullIndividualBuffs,
		Rotation:        core.GetAplRotation("../ui/protection_warrior/apls", "default").Rotation,
		InFrontOfTarget: true,
	}, core.FullPartyBuffs, core.FullRaidBuffs, core.FullDebuffs)
	raid.Tanks = []*proto.UnitReference{{Type: proto.UnitReference_Player, Index: 0}}
	return &proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  &proto.Encounter{Duration: 300, Targets: []*proto.Target{preset.Config}},
		SimOptions: &proto.SimOptions{Iterations: 5, RandomSeed: 41, Debug: true},
	}
}

func TestBossMeleeWaitsWhileItCasts(t *testing.T) {
	for _, tc := range []struct {
		path      string
		hardcasts int
	}{
		{"ToGC 25/Gormok", 20},
		{"Ulduar 25/Hodir", 20},
	} {
		result := core.RunRaidSim(bossCastRequest(t, tc.path))
		if result.ErrorResult != "" {
			t.Fatalf("%s: %s", tc.path, result.ErrorResult)
		}
		c := countBossCasts(t, result.Logs)
		t.Logf("%s: %+v", tc.path, c)
		if c.iterations != 5 || c.hardcasts < tc.hardcasts || c.swings < 300 {
			t.Fatalf("%s: %d iterations, %d boss hardcasts, %d boss swings in the log; want 5, at least %d and at least 300", tc.path, c.iterations, c.hardcasts, c.swings, tc.hardcasts)
		}
		if c.inside != 0 {
			t.Errorf("%s: %d boss swings landed inside its own cast, want 0", tc.path, c.inside)
		}
		if c.completed+c.running != c.hardcasts || c.running > c.iterations || c.late != 0 || c.ended != 0 || c.cutOff != 0 {
			t.Errorf("%s: of %d boss hardcasts %d completed (%d late), %d ran out without completing, %d were cut off by another cast, %d were still running when the log ends; want every cast completed on time", tc.path, c.hardcasts, c.completed, c.late, c.ended, c.cutOff, c.running)
		}
	}
}

func TestBossCastCompletesWithoutASwing(t *testing.T) {
	for _, speed := range []float64{0, 12} {
		rsr := bossCastRequest(t, "Ulduar 25/Hodir")
		target := googleProto.Clone(rsr.Encounter.Targets[0]).(*proto.Target)
		target.SwingSpeed = speed
		rsr.Encounter.Targets[0] = target
		result := core.RunRaidSim(rsr)
		if result.ErrorResult != "" {
			t.Fatalf("Hodir, swing speed %v: %s", speed, result.ErrorResult)
		}
		c := countBossCasts(t, result.Logs)
		t.Logf("Hodir, swing speed %v: %+v", speed, c)
		if c.iterations != 5 || c.hardcasts < 20 {
			t.Fatalf("Hodir, swing speed %v: %d iterations, %d Flash Freezes in the log; want 5 and at least 20", speed, c.iterations, c.hardcasts)
		}
		if c.completed != c.hardcasts || c.late != 0 || c.inside != 0 {
			t.Errorf("Hodir, swing speed %v: of %d Flash Freezes %d completed (%d late), %d ran out without completing, %d cut off, %d still running, %d swings inside; want every one completed on time", speed, c.hardcasts, c.completed, c.late, c.ended, c.cutOff, c.running, c.inside)
		}
	}
}
