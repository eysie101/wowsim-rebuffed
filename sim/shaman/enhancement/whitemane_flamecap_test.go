package enhancement

import (
	"strings"
	"testing"

	"github.com/wowsims/wotlk/sim/core"
)

func TestWhitemaneFlameCapProcsOnWhiteSwings(t *testing.T) {
	const own = "] [Enha (#1)] "
	windows, whites, procs, strikes, misplaced, afterStrike := 0, 0, 0, 0, 0, 0
	for seed := int64(1); seed <= 5; seed++ {
		rsr := whitemaneRequest(1)
		rsr.SimOptions.RandomSeed = seed
		rsr.SimOptions.Debug = true
		result := core.RunRaidSim(rsr)
		if result.ErrorResult != "" {
			t.Fatal(result.ErrorResult)
		}
		active, start, white, proc, last := false, "", 0, 0, ""
		for _, line := range strings.Split(result.Logs, "\n") {
			if !strings.Contains(line, own) || strings.Contains(line, "[DEBUG]") {
				continue
			}
			stamp := strings.SplitN(line, " ", 2)[0]
			landed := strings.Contains(line, " damage.") && !strings.Contains(line, "for 0.000 damage")
			switch {
			case strings.Contains(line, "Aura gained: {ItemID: 22788}"):
				active, start, white, proc, last = true, stamp, 0, 0, ""
				windows++
			case strings.Contains(line, "Aura faded: {ItemID: 22788}"):
				if active && proc != white {
					t.Errorf("seed %d, Flame Cap %s to %s: %d Flamecap Fire procs for %d landed white swings", seed, start, stamp, proc, white)
				}
				active = false
			case !strings.Contains(line, own+"[Target"):
			case strings.Contains(line, "{ItemID: 22788}"):
				if !active {
					t.Errorf("seed %d: Flamecap Fire without Flame Cap: %s", seed, line)
				}
				proc++
				procs++
				if last != "white" {
					misplaced++
				}
				if last == "strike" {
					afterStrike++
				}
				last = ""
			case strings.Contains(line, "{OtherID: 3, Tag: 1}") || strings.Contains(line, "{OtherID: 3, Tag: 2}"):
				last = ""
				if active && landed {
					white++
					whites++
					last = "white"
				}
			case strings.Contains(line, "{SpellID: 17364}") || strings.Contains(line, "{SpellID: 60103}"):
				last = ""
				if active && landed {
					strikes++
					last = "strike"
				}
			}
		}
	}
	if windows < 10 || whites < 500 || strikes < 50 {
		t.Fatalf("parser not calibrated: %d Flame Cap windows, %d landed white swings and %d landed Stormstrike or Lava Lash strikes in them", windows, whites, strikes)
	}
	if misplaced > 0 {
		t.Errorf("%d of %d Flamecap Fire procs did not follow a landed white swing, %d of them a Stormstrike or Lava Lash strike", misplaced, procs, afterStrike)
	}
	t.Logf("%d Flame Cap windows: %d landed white swings, %d Flamecap Fire procs, %d landed Stormstrike or Lava Lash strikes", windows, whites, procs, strikes)
}
