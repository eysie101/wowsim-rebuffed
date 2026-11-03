package enhancement

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func TestWhitemaneFireNovaNeedsFireTotem(t *testing.T) {
	sim := core.NewSim(whitemaneRequest(1))
	sim.Reset()
	enh := sim.Raid.Parties[0].Players[0].(*EnhancementShaman)
	target := enh.CurrentTarget
	check := func(when string, want bool) {
		if got := enh.FireNova.CanCast(sim, target); got != want {
			t.Errorf("%s at %s: Fire Nova castable %v, want %v", when, sim.CurrentTime, got, want)
		}
	}
	check("no fire totem at the start", false)
	enh.MagmaTotem.SkipCastAndApplyEffects(sim, target)
	check("Magma Totem down", true)
	sim.CurrentTime = 22 * time.Second
	check("Magma Totem expired", false)
	fire := enh.Totems.Fire
	defer func() { enh.Totems.Fire = fire }()
	enh.Totems.Fire = proto.FireTotem_NoFireTotem
	enh.FireElementalTotem.SkipCastAndApplyEffects(sim, target)
	check("Fire Elemental Totem down, no fire totem configured", true)
	sim.Cleanup()
	sim.Reset()
	check("next iteration, no fire totem yet", false)
}

func whitemanePrepullCast(spellID int32, at string) *proto.APLPrepullAction {
	return &proto.APLPrepullAction{
		Action:    &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: spellID}}}}},
		DoAtValue: &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: at}}},
	}
}

func TestWhitemaneFireNovaBeforeThePullAfterAPrepullTotem(t *testing.T) {
	rsr := whitemaneRequest(2)
	rsr.SimOptions.Debug = true
	rotation := rsr.Raid.Parties[0].Players[0].Rotation
	rotation.PrepullActions = append(rotation.PrepullActions, whitemanePrepullCast(58734, "-3s"), whitemanePrepullCast(61657, "-1s"))
	result := core.RunRaidSim(rsr)
	if result.ErrorResult != "" {
		t.Fatal(result.ErrorResult)
	}
	if novas := strings.Count(result.Logs, "[-1.00] [Enha (#1)] Completed cast {SpellID: 61657}"); novas != 2 {
		t.Errorf("%d Fire Novas at -1s after a Magma Totem at -3s in 2 iterations, want 2", novas)
	}
}

func TestWhitemaneFireNovaAboveMagmaWaitsForATotem(t *testing.T) {
	novas, prepull, bad := 0, 0, 0
	for _, seed := range []int64{21, 22, 23} {
		rsr := whitemaneRequest(2)
		rsr.SimOptions.RandomSeed = seed
		rsr.SimOptions.Debug = true
		rotation := rsr.Raid.Parties[0].Players[0].Rotation
		fn, mag := -1, -1
		for i, item := range rotation.PriorityList {
			switch item.GetAction().GetCastSpell().GetSpellId().GetSpellId() {
			case 61657:
				fn = i
			case 58734:
				mag = i
			}
		}
		if fn < 0 || mag < 0 || fn < mag {
			t.Fatalf("enha_p1_spell: Fire Nova at %d, Magma Totem at %d, want both with Magma first", fn, mag)
		}
		list := append([]*proto.APLListItem{}, rotation.PriorityList[:mag]...)
		list = append(list, rotation.PriorityList[fn])
		list = append(list, rotation.PriorityList[mag:fn]...)
		rotation.PriorityList = append(list, rotation.PriorityList[fn+1:]...)
		rotation.PrepullActions = append(rotation.PrepullActions, whitemanePrepullCast(61657, "-10s"))
		result := core.RunRaidSim(rsr)
		if result.ErrorResult != "" {
			t.Fatal(result.ErrorResult)
		}
		magma, elemental, previous, iterations := -1000.0, -1000.0, 1e9, 0
		for _, line := range strings.Split(result.Logs, "\n") {
			if !strings.Contains(line, "] [Enha (#1)] Completed cast {SpellID: ") {
				continue
			}
			at, err := strconv.ParseFloat(strings.Trim(strings.SplitN(line, " ", 2)[0], "[]"), 64)
			if err != nil {
				t.Fatal(err)
			}
			if at < previous {
				magma, elemental = -1000, -1000
				iterations++
			}
			previous = at
			if at < 0 && strings.Contains(line, "{SpellID: 61657}") {
				prepull++
			}
			switch {
			case strings.Contains(line, "{SpellID: 58734}"):
				magma = at
			case strings.Contains(line, "{SpellID: 2894}"):
				elemental = at
			case strings.Contains(line, "{SpellID: 61657}"):
				novas++
				if at-magma > 20 && at-elemental > 120 {
					bad++
					t.Logf("seed %d iteration %d: Fire Nova at %.2fs, last Magma Totem at %.2fs, last Fire Elemental Totem at %.2fs", seed, iterations, at, magma, elemental)
				}
			}
		}
		if iterations != 2 {
			t.Fatalf("seed %d: %d iterations in the log, want 2", seed, iterations)
		}
	}
	t.Logf("%d Fire Novas, %d before the pull, %d without a fire totem", novas, prepull, bad)
	if novas == 0 || bad > 0 {
		t.Errorf("%d Fire Novas, %d of them without a fire totem: want Fire Novas, none without a totem", novas, bad)
	}
}
