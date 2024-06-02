package game

import (
	"gladiatorsGoModule/gameJson"
	"matchgame/packet"
	// log "github.com/sirupsen/logrus"
	// "gladiatorsGoModule/utility"
	// "matchgame/logger"
)

type Gladiator struct {
	ID              string // DBGladiator的_id
	LeftSide        bool   // true是左方玩家(第一位), false是右方玩家(第二位)
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
	CurUnit         int
	Speed           int
}

func NewGladiator(id string, jsonGladiator gameJson.JsonGladiator, jsonSkills [GladiatorSkillCount]gameJson.JsonSkill,
	jsonTraits []gameJson.TraitJsonData, jsonEquips []gameJson.JsonEquip) (Gladiator, error) {
	pos := -InitGladiatorPos
	leftSide := true
	if len(MyRoom.Gamers) == 1 {
		pos = InitGladiatorPos
		leftSide = false
	}

	gladiator := Gladiator{
		ID:            id,
		LeftSide:      leftSide,
		JsonGladiator: jsonGladiator,
		JsonSkills:    jsonSkills,
		JsonTraits:    jsonTraits,
		JsonEquips:    jsonEquips,
		CurUnit:       pos,
		Speed:         jsonGladiator.Speed,
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
		JsonGladiatorID: gladiator.JsonGladiator.ID,
		JsonSkillIDs:    jsonSkillIDs,
		JsonTraitIDs:    jsonTraitIDs,
		JsonEquipIDs:    jsonEquipIDs,
		CurSkills:       curSkills,
		HP:              gladiator.HP,
		CurHP:           gladiator.CurHP,
		CurVigor:        gladiator.CurVigor,
		Speed:           gladiator.Speed,
	}
	return packGladiator
}

func (gladiator *Gladiator) Dir() int {
	if gladiator.LeftSide {
		return 1
	} else {
		return -1
	}
}

func (gladiator *Gladiator) Move() {
	gladiator.CurUnit += gladiator.Speed * gladiator.Dir()
}

func (gladiator *Gladiator) CurGrid() int {
	return gladiator.CurUnit / GridUnit
}

func (gladiator *Gladiator) DmgBuff() float64 {
	return float64(gladiator.STR) / float64(100)
}
