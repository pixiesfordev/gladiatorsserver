package game

import (
	"gladiatorsGoModule/gameJson"
	"matchgame/packet"

	log "github.com/sirupsen/logrus"
	// log "github.com/sirupsen/logrus"
	// "gladiatorsGoModule/utility"
	// "matchgame/logger"
)

type Gladiator struct {
	ID              string // DBGladiator的_id
	JsonGladiator   gameJson.JsonGladiator
	JsonSkills      [GladiatorSkillCount]gameJson.JsonSkill
	JsonTraits      []gameJson.TraitJsonData
	JsonEquips      []gameJson.JsonEquip
	CurJsonSkillIDs [4]packet.PackSkill
	HP              int
	CurHP           int
	CurVigor        float64
	VigorRegon      float64
	STR             int
	DEF             int
	MDEF            int
	CRIT            float64
	INIT            int
	Knockback       int
	GladiatorPos
}

func NewGladiator(id string, jsonGladiator gameJson.JsonGladiator, jsonSkills [GladiatorSkillCount]gameJson.JsonSkill,
	jsonTraits []gameJson.TraitJsonData, jsonEquips []gameJson.JsonEquip) (Gladiator, error) {
	pos := -InitGladiatorPos
	leftSide := true
	log.Infof("NewGladiator when myRoom: %v, %d", MyRoom.Gamers, len(MyRoom.Gamers))
	if len(MyRoom.Gamers) >= 1 {
		pos = InitGladiatorPos
		leftSide = false
	}

	gladiator := Gladiator{
		ID:            id,
		JsonGladiator: jsonGladiator,
		JsonSkills:    jsonSkills,
		JsonTraits:    jsonTraits,
		JsonEquips:    jsonEquips,
		Knockback:     jsonGladiator.Knockback,
		GladiatorPos: GladiatorPos{
			LeftSide: leftSide,
			CurUnit:  pos,
			Speed:    jsonGladiator.Speed,
		},
	}
	return gladiator, nil
}

func (gladiator *Gladiator) GetPackGladiator() packet.PackGladiator {
	var jsonSkillIDs [GladiatorSkillCount]int
	jsonTraitIDs := make([]int, 0)
	jsonEquipIDs := make([]int, 0)
	var curSkills [4]packet.PackSkill
	for i, _ := range curSkills {
		curSkills[i] = packet.PackSkill{
			JsonID: gladiator.CurJsonSkillIDs[i].JsonID,
			On:     gladiator.CurJsonSkillIDs[i].On,
		}
	}

	packGladiator := packet.PackGladiator{
		LeftSide:        gladiator.LeftSide, // 之後可刪
		JsonGladiatorID: gladiator.JsonGladiator.ID,
		JsonSkillIDs:    jsonSkillIDs,
		JsonTraitIDs:    jsonTraitIDs,
		JsonEquipIDs:    jsonEquipIDs,
		CurSkills:       curSkills,
		HP:              gladiator.HP,
		CurHP:           gladiator.CurHP,
		CurVigor:        gladiator.CurVigor,
		Knockback:       gladiator.Knockback,
		Speed:           gladiator.Speed + gladiator.Rush,
		BattlePos:       gladiator.CurUnit,
		StagePos:        gladiator.CurGrid(),
		Rush:            gladiator.Rush > 0,
	}
	return packGladiator
}

func (gladiator *Gladiator) DmgBuff() float64 {
	return float64(gladiator.STR) / float64(100)
}
