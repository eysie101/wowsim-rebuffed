package enhancement

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/wotlk/sim/core"
)

var whitemaneCastLine = regexp.MustCompile(`^\[(-?\d+\.\d+)\] \[Enha \(#1\)\] (?:Casting (\{[^}]*\}) \(Cost = [^,]*, Cast Time = ([^,]*),|Completed cast (\{[^}]*\}))`)

type whitemaneCastCount struct {
	iterations, hardcasts, swings, inside, late, dropped int
}

func whitemaneCountCasts(t *testing.T, logs string) whitemaneCastCount {
	type cast struct {
		id         string
		start, end float64
	}
	var c whitemaneCastCount
	var pending *cast
	instant, previous := "", 1e9
	for _, line := range strings.Split(logs, "\n") {
		m := whitemaneCastLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		at, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			t.Fatal(err)
		}
		if at < previous-1 {
			c.iterations++
			pending, instant = nil, ""
		}
		previous = at
		if m[2] != "" {
			d, err := time.ParseDuration(m[3])
			if err != nil {
				t.Fatal(err)
			}
			if d == 0 {
				instant = m[2]
				if strings.HasPrefix(m[2], "{OtherID: 3, ") {
					c.swings++
					if pending != nil && at > pending.start+0.011 && at < pending.end-0.011 {
						c.inside++
					}
				}
				continue
			}
			if pending != nil {
				c.dropped++
			}
			c.hardcasts++
			pending = &cast{m[2], at, at + d.Seconds()}
			continue
		}
		if m[4] == instant {
			instant = ""
			continue
		}
		if pending != nil && m[4] == pending.id {
			if at > pending.end+0.015 {
				c.late++
			}
			pending = nil
		}
	}
	return c
}

func TestWhitemaneHardcastHoldsMelee(t *testing.T) {
	for _, seed := range []int64{31, 32} {
		rsr := whitemaneRequest(5)
		rsr.SimOptions.RandomSeed = seed
		rsr.SimOptions.Debug = true
		result := core.RunRaidSim(rsr)
		if result.ErrorResult != "" {
			t.Fatal(result.ErrorResult)
		}
		c := whitemaneCountCasts(t, result.Logs)
		t.Logf("seed %d: %+v", seed, c)
		if c.iterations != 5 || c.hardcasts < 200 || c.swings < 1000 {
			t.Fatalf("seed %d: %d iterations, %d hardcasts, %d swings in the log; want 5 iterations, at least 200 hardcasts and 1000 swings", seed, c.iterations, c.hardcasts, c.swings)
		}
		if c.inside != 0 {
			t.Errorf("seed %d: %d melee swings landed inside a hardcast, want 0", seed, c.inside)
		}
		if c.late != 0 || c.dropped != 0 {
			t.Errorf("seed %d: %d hardcasts completed late and %d never completed, want 0 and 0", seed, c.late, c.dropped)
		}
	}
}
