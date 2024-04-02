package game

import (
	"gladiatorsGoModule/gameJson"
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
	HP            int64
	CurHP         int64
	VigorRegon    float64
	STR           int64
	DEF           int64
	MDEF          int64
	CRIT          float64
	INIT          int64
	Knockback     int64
}
