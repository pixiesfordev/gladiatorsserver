package game

import (
	"gladiatorsGoModule/setting"
	"matchgame/packet"

	log "github.com/sirupsen/logrus"
)

const (
	GladiatorRush = 4
	GirdUnit      = 1000
	TimeUnit      = 10
	TimeMili      = 1000
)

type GladiatorPos struct {
	LeftSide      bool
	CurUnit       float64
	Speed         float64
	Rush          float64
	CantMoveTimer float64
}

func (gPos *GladiatorPos) Dir() float64 {
	if gPos.LeftSide {
		return 1
	} else {
		return -1
	}
}

func (gPos *GladiatorPos) SetRush(on bool, rush float64) {
	if on && gPos.Rush == 0 {
		gPos.Rush += rush
	} else if !on && gPos.Rush >= 0 {
		gPos.Rush = 0
	}
}

func (gPos *GladiatorPos) CanMove() bool {
	return gPos.CantMoveTimer <= 0
}

func (gPos *GladiatorPos) AddCantMoveTimer(secs float64) {
	gPos.CantMoveTimer += secs
}

func (gPos *GladiatorPos) MoveAhead(d float64) {
	gPos.CurUnit += d * gPos.Dir()
	if gPos.CurUnit > WallPos {
		gPos.CurUnit = WallPos
	} else if gPos.CurUnit < -WallPos {
		gPos.CurUnit = -WallPos
	}
}

func (gPos *GladiatorPos) MoveUnitByTime(secs float64) bool { // 毫秒為單位
	if gPos.CantMoveTimer > 0 {
		gPos.CantMoveTimer -= secs
		return false
	}
	totalSpeed := gPos.Speed + gPos.Rush
	gPos.MoveAhead(totalSpeed * secs)
	return true
}

// 擊退xUnit,
func (gPos *GladiatorPos) KnockBackUnitByTime(unit float64, secs float64) {
	gPos.AddCantMoveTimer(secs)
	gPos.MoveAhead(-unit)
}

func IsCollide() bool {
	dis := LeftGamer.GetGladiator().CurUnit - RightGamer.GetGladiator().CurUnit
	if dis >= 0 {
		return dis <= CollisionDis
	} else {
		return -dis <= CollisionDis
	}
}

func GetCollisionData() ([setting.PLAYER_NUMBER]packet.PackPlayerState, float64) {
	collisionPos := 0 /*(LeftGamer.GetGladiator().CurUnit + RightGamer.GetGladiator().CurUnit) / 2*/
	LeftBack := float64(RightGamer.GetGladiator().Knockback)
	RightBack := float64(LeftGamer.GetGladiator().Knockback)
	LeftGamer.GetGladiator().KnockBackUnitByTime(LeftBack, KNOCK_BACK_SECS*TimeMili)
	RightGamer.GetGladiator().KnockBackUnitByTime(RightBack, KNOCK_BACK_SECS*TimeMili)
	log.Infof("GetCollision: GameTime(%f, %d) End with POS(%d ,%d), collisionPos: %d, BackDis(%d, %d), Speed(%d, %d), Rush(%d, %d)",
		float64(GameTime)/float64(TimeMili),
		GameTime,
		LeftGamer.GetGladiator().CurUnit, RightGamer.GetGladiator().CurUnit,
		collisionPos,
		LeftBack, RightBack,
		LeftGamer.GetGladiator().Speed, RightGamer.GetGladiator().Speed,
		LeftGamer.GetGladiator().Rush, RightGamer.GetGladiator().Rush,
	)
	return MyRoom.GetPackPlayerStates(), GameTime + KNOCK_BACK_SECS
}
