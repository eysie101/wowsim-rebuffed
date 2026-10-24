package enhancement

import (
	"strings"
	"testing"

	"github.com/wowsims/wotlk/sim/core"
)

func TestWhitemaneStormstrikeFlametongueProcs(t *testing.T) {
	instants, bad, seeds := 0, 0, []int64{11, 12, 13, 14, 15}
	for _, seed := range seeds {
		rsr := whitemaneRequest(1)
		rsr.SimOptions.RandomSeed = seed
		rsr.SimOptions.Debug = true
		result := core.RunRaidSim(rsr)
		if result.ErrorResult != "" {
			t.Fatal(result.ErrorResult)
		}
		groups := map[string][]string{}
		order := []string{}
		for _, line := range strings.Split(result.Logs, "\n") {
			if !strings.Contains(line, "] [Enha (#1)] [Target") || strings.Contains(line, "[DEBUG]") {
				continue
			}
			stamp := strings.SplitN(line, " ", 2)[0]
			if _, seen := groups[stamp]; !seen {
				order = append(order, stamp)
			}
			groups[stamp] = append(groups[stamp], line)
		}
		for _, stamp := range order {
			lines := groups[stamp]
			strikes, procs, other := 0, 0, false
			for _, l := range lines {
				switch {
				case strings.Contains(l, "{SpellID: 17364}") && strings.Contains(l, " damage") && !strings.Contains(l, "for 0.000 damage"):
					strikes++
				case strings.Contains(l, "{SpellID: 58790}") || strings.Contains(l, "{SpellID: 58789}"):
					procs++
				case strings.Contains(l, "OtherID:") || strings.Contains(l, "{SpellID: 60103}") || strings.Contains(l, "{SpellID: 25504}"):
					other = true
				}
			}
			if strikes != 2 || other {
				continue
			}
			instants++
			if procs != 2 {
				bad++
				if bad <= 5 {
					t.Errorf("seed %d, instant %s: %d Flametongue procs for 2 landed strikes:\n%s", seed, stamp, procs, strings.Join(lines, "\n"))
				}
			}
		}
	}
	if instants < 50 {
		rsr := whitemaneRequest(1)
		rsr.SimOptions.Debug = true
		sample, n := []string{}, 0
		for _, line := range strings.Split(core.RunRaidSim(rsr).Logs, "\n") {
			n++
			if strings.Contains(line, "17364") && len(sample) < 12 {
				sample = append(sample, line)
			}
		}
		t.Fatalf("only %d isolated two-strike Stormstrike instants found in %d debug iterations; the gate needs at least 50. Log lines %d, sample:\n%s", instants, len(seeds), n, strings.Join(sample, "\n"))
	}
	if bad > 0 {
		t.Fatalf("%d of %d isolated two-strike Stormstrikes did not give exactly 2 Flametongue procs", bad, instants)
	}
	t.Logf("%d isolated two-strike Stormstrikes, all with exactly 2 Flametongue procs", instants)
}
