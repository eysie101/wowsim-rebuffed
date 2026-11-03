package dps

import (
	"strconv"
	"testing"
	"time"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func TestWhitemaneSweepingStrikesTenAttacks(t *testing.T) {
	rsr := armsRequest(1, 2)
	rsr.SimOptions.Debug = false
	rsr.SimOptions.IsTest = true
	sim := core.NewSim(rsr)
	sim.Reset()
	war := sim.Raid.Parties[0].Players[0].(*DpsWarrior).Warrior
	ss := war.GetAura("Sweeping Strikes")
	ssHit := war.GetSpell(core.ActionID{SpellID: 12723})
	if ss == nil || ssHit == nil {
		t.Fatal("no Sweeping Strikes on the Arms warrior")
	}
	copies := func() int32 {
		n := int32(0)
		for _, target := range sim.Encounter.TargetUnits {
			n += ssHit.SpellMetrics[target.UnitIndex].Hits
		}
		return n
	}
	var start int32
	var windows []int32
	onGain, onExpire := ss.OnGain, ss.OnExpire
	ss.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		onGain(aura, sim)
		start = copies()
	}
	ss.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		if onExpire != nil {
			onExpire(aura, sim)
		}
		if aura.GetStacks() == 0 {
			windows = append(windows, copies()-start)
		}
	}
	sim.PrePull()
	for !sim.Step() {
	}
	if len(windows) < 5 {
		t.Fatalf("only %d Sweeping Strikes used up their charges in 300 s against two targets", len(windows))
	}
	for i, n := range windows {
		if n != 10 {
			t.Errorf("Sweeping Strikes %d of %d struck an additional target %d times before it faded, want 10", i+1, len(windows), n)
		}
	}
}

func TestWhitemaneImprovedMortalStrike(t *testing.T) {
	var untalented float64
	for rank, bonus := range []float64{0, 0.12, 0.23, 0.35} {
		rsr := armsRequest(1, 1)
		rsr.SimOptions.Debug = false
		rsr.SimOptions.IsTest = true
		rsr.Raid.Debuffs = &proto.Debuffs{}
		player := rsr.Raid.Parties[0].Players[0]
		player.TalentsString = ArmsTalents[:24] + strconv.Itoa(rank) + ArmsTalents[25:]
		player.Glyphs = &proto.Glyphs{}
		sim := core.NewSim(rsr)
		sim.Reset()
		war := sim.Raid.Parties[0].Players[0].(*DpsWarrior).Warrior
		if war.Talents.ImprovedMortalStrike != int32(rank) || !war.Talents.MortalStrike {
			t.Fatalf("talents %s: Mortal Strike %v, Improved Mortal Strike %d; want Mortal Strike and rank %d", player.TalentsString, war.Talents.MortalStrike, war.Talents.ImprovedMortalStrike, rank)
		}
		if cd := war.MortalStrike.CD.Duration; cd != 6*time.Second {
			t.Errorf("Improved Mortal Strike %d/3: Mortal Strike cooldown %s, want 6s", rank, cd)
		}
		damage := war.MortalStrike.CalcDamage(sim, sim.Encounter.TargetUnits[0], 1000, war.MortalStrike.OutcomeAlwaysHit).Damage
		if rank == 0 {
			untalented = damage
			continue
		}
		if got := damage / untalented; !core.WithinToleranceFloat64(1+bonus, got, 1e-9) {
			t.Errorf("Improved Mortal Strike %d/3: Mortal Strike deals %.4f times the untalented damage, want %.2f", rank, got, 1+bonus)
		}
	}
}
