package packet

import (
// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"
)

type Melee_ToClient struct {
	CMDContent
	MyPlayerState       PackPlayerState
	OpponentPlayerState PackPlayerState
	MyAttack            PackAttack
	OpponentAttack      PackAttack
	GameTime            float64
}

type PackAttack struct {
	AttackPos float64
	Knockback float64
	SkillID   int
}
