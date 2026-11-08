package enhancement

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func TestWhitemaneSundialTriggers(t *testing.T) {
	const player = "] [Enha (#1)] "
	const fights = 10
	triggers := map[string]int{}
	procs, flametongues := 0, 0
	for seed := int64(1); seed <= fights; seed++ {
		rsr := whitemaneRequest(1)
		rsr.SimOptions.RandomSeed = seed
		rsr.SimOptions.Debug = true
		rsr.Raid.Parties[0].Players[0].Equipment.Items[proto.ItemSlot_ItemSlotTrinket1].Id = 40682
		result := core.RunRaidSim(rsr)
		if result.ErrorResult != "" {
			t.Fatal(result.ErrorResult)
		}
		stamp, last := "", ""
		for _, line := range strings.Split(result.Logs, "\n") {
			if !strings.Contains(line, player) {
				continue
			}
			if at := strings.SplitN(line, " ", 2)[0]; at != stamp {
				stamp, last = at, ""
			}
			if _, cast, found := strings.Cut(line, player+"Completed cast "); found {
				last = cast
				if cast == "{SpellID: 58790}" || cast == "{SpellID: 58789}" {
					flametongues++
				}
			} else if strings.Contains(line, player+"Aura gained: {SpellID: 60064}") {
				procs++
				triggers[last]++
			}
		}
	}
	casts := make([]string, 0, len(triggers))
	for cast := range triggers {
		casts = append(casts, cast)
	}
	sort.Strings(casts)
	table := make([]string, 0, len(casts))
	for _, cast := range casts {
		table = append(table, fmt.Sprintf("%q %d", cast, triggers[cast]))
	}
	t.Logf("%d Sundial procs and %d Flametongue Attacks in %d fights; procs by the cast completed before them: %s", procs, flametongues, fights, strings.Join(table, ", "))
	if procs < 2*fights || flametongues < 100*fights {
		t.Fatalf("parser not calibrated: %d Sundial procs and %d Flametongue Attacks in %d fights", procs, flametongues, fights)
	}
	if n := triggers["{SpellID: 58790}"] + triggers["{SpellID: 58789}"]; n == 0 {
		t.Fatalf("none of %d Sundial procs came from one of %d Flametongue Attacks", procs, flametongues)
	}
	if triggers["{SpellID: 49279}"] == 0 {
		t.Fatalf("none of %d Sundial procs came from a Lightning Shield orb", procs)
	}
	if triggers["{ItemID: 22788}"] == 0 {
		t.Fatalf("none of %d Sundial procs came from Flamecap Fire", procs)
	}
}
