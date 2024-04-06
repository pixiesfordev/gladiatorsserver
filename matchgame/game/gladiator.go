package game

import (
	"gladiatorsGoModule/gameJson"
	"matchgame/packet"
	// log "github.com/sirupsen/logrus"
	// "gladiatorsGoModule/utility"
	// "matchgame/logger"
)

type Gladiator struct {
	ID            string // DBGladiator的_id
	JsonGladiator gameJson.JsonGladiator
	JsonSkills    [6]gameJson.JsonSkill
	JsonTraits    []gameJson.TraitJsonData
	JsonEquips    []gameJson.JsonEquip
	HP            int
	CurHP         int
	CurVigor      float64
	VigorRegon    float64
	STR           int
	DEF           int
	MDEF          int
	CRIT          float64
	INIT          int
	Knockback     int
}

func (gladiator *Gladiator) GetPackGladiator() packet.PackGladiator {

	var jsonSkillIDs [6]int
	jsonTraitIDs := make([]int, 0)
	jsonEquipIDs := make([]int, 0)

	packGladiator := packet.PackGladiator{
		JsonGladiatorID: gladiator.JsonGladiator.ID,
		JsonSkillIDs:    jsonSkillIDs,
		JsonTraitIDs:    jsonTraitIDs,
		JsonEquipIDs:    jsonEquipIDs,
		HP:              gladiator.HP,
		CurHP:           gladiator.CurHP,
		CurVigor:        gladiator.CurVigor,
	}
	return packGladiator
}
