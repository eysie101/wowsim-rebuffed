package tank

import (
	"testing"

	googleProto "google.golang.org/protobuf/proto"

	"github.com/wowsims/wotlk/sim/core"
	"github.com/wowsims/wotlk/sim/core/proto"
)

func TestWhitemaneFlameCapNotOnRuneStrike(t *testing.T) {
	player := &proto.Player{
		Name:          "DK",
		Race:          proto.Race_RaceOrc,
		Class:         proto.Class_ClassDeathknight,
		Equipment:     core.GetGearSet("../../../ui/tank_deathknight/gear_sets", "p1_blood").GearSet,
		Consumes:      googleProto.Clone(FullConsumes).(*proto.Consumes),
		Spec:          PlayerOptionsBloodTank,
		Glyphs:        Glyphs,
		TalentsString: BloodTankTalents,
		Buffs:         core.FullIndividualBuffs,
		Rotation:      core.GetAplRotation("../../../ui/tank_deathknight/apls", "blood_icy_touch").Rotation,
	}
	player.Consumes.DefaultConjured = proto.Conjured_ConjuredFlameCap
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       core.SinglePlayerRaidProto(player, core.FullPartyBuffs, core.FullRaidBuffs, core.FullDebuffs),
		Encounter:  &proto.Encounter{Duration: 300, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1},
	})
	sim.Reset()
	char := sim.Raid.Parties[0].Players[0].GetCharacter()
	aura := char.GetAuraByID(core.ActionID{ItemID: 22788})
	var fire *core.Spell
	for _, spell := range char.Spellbook {
		if spell.ActionID.SameAction(core.ActionID{ItemID: 22788}) && spell.ProcMask == core.ProcMaskEmpty {
			fire = spell
		}
	}
	if aura == nil || fire == nil {
		t.Fatalf("Flame Cap aura found %v, Flamecap Fire found %v", aura != nil, fire != nil)
	}
	aura.Activate(sim)
	target := char.CurrentTarget

	runeStrikes, autoProcs := 0, 0
	for _, spell := range char.Spellbook {
		before := fire.SpellMetrics[target.UnitIndex].Casts
		result := spell.NewResult(target)
		result.Outcome = core.OutcomeHit
		aura.OnSpellHitDealt(aura, sim, spell, result)
		procced := fire.SpellMetrics[target.UnitIndex].Casts > before
		switch {
		case spell.ActionID.SpellID == 56815:
			runeStrikes++
			if procced {
				t.Errorf("a landed %s procced Flamecap Fire", spell.ActionID)
			}
		case spell.ActionID.SameAction(core.ActionID{OtherID: proto.OtherAction_OtherActionAttack, Tag: 1}):
			if !procced {
				t.Errorf("a landed main-hand auto attack %s did not proc Flamecap Fire", spell.ActionID)
			}
			autoProcs++
		case procced && !spell.ProcMask.Matches(core.ProcMaskWhiteHit):
			t.Errorf("%s has no auto attack bit and procced Flamecap Fire", spell.ActionID)
		}
	}
	if runeStrikes < 2 || autoProcs != 1 {
		t.Fatalf("setup not calibrated: %d Rune Strike spells (want main and off hand), %d main-hand auto attacks", runeStrikes, autoProcs)
	}
}
