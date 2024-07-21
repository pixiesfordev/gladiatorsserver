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
	CurUnit       int
	Speed         int // grid/secs
	Rush          int // grid/secs
	CantMoveTimer int // milisecs
}

func (gPos *GladiatorPos) Dir() int {
	if gPos.LeftSide {
		return 1
	} else {
		return -1
	}
}

func (gPos *GladiatorPos) SetRush(on bool, rush int) {
	if on && gPos.Rush == 0 {
		gPos.Rush += rush
	} else if !on && gPos.Rush >= 0 {
		gPos.Rush = 0
	}
}

func (gPos *GladiatorPos) CanMove() bool {
	return gPos.CantMoveTimer <= 0
}

func (gPos *GladiatorPos) AddCantMoveTimer(milisecs int) {
	gPos.CantMoveTimer += milisecs
}

func (gPos *GladiatorPos) MoveAhead(d int) {
	gPos.CurUnit += d * gPos.Dir()
	if gPos.CurUnit > WallPos {
		gPos.CurUnit = WallPos
	} else if gPos.CurUnit < -WallPos {
		gPos.CurUnit = -WallPos
	}
}

func (gPos *GladiatorPos) MoveUnitByTime(milisecs int) bool { // 毫秒為單位
	if gPos.CantMoveTimer > 0 {
		gPos.CantMoveTimer -= milisecs
		return false
	}
	totalSpeed := gPos.Speed + gPos.Rush
	gPos.MoveAhead(totalSpeed * GridUnit / TimeMili * milisecs)
	return true
}

// 擊退xUnit,
func (gPos *GladiatorPos) KnockBackUnitByTime(Unit int, milisecs int) {
	gPos.AddCantMoveTimer(milisecs)
	gPos.MoveAhead(-Unit)
}

func (gPos *GladiatorPos) CurGrid() float64 {
	return float64(gPos.CurUnit) / float64(GridUnit)
}

func IsCollide() bool {
	dis := LeftGamer.GetGladiator().CurUnit - RightGamer.GetGladiator().CurUnit
	if dis >= 0 {
		return dis <= CollisionDis*GridUnit
	} else {
		return -dis <= CollisionDis*GridUnit
	}
}

func GetCollisionData() ([setting.PLAYER_NUMBER]packet.PackPlayerState, int) {
	collisionPos := 0 /*(LeftGamer.GetGladiator().CurUnit + RightGamer.GetGladiator().CurUnit) / 2*/
	LeftBack := /*LeftGamer.GetGladiator().CurUnit - collisionPos +*/ (RightGamer.GetGladiator().Knockback * GridUnit)
	RightBack := /*collisionPos - RightGamer.GetGladiator().CurUnit +*/ (LeftGamer.GetGladiator().Knockback * GridUnit)
	LeftGamer.GetGladiator().KnockBackUnitByTime(LeftBack, KNOCK_BACK_TIME*TimeMili)
	RightGamer.GetGladiator().KnockBackUnitByTime(RightBack, KNOCK_BACK_TIME*TimeMili)
	log.Infof("GetCollision: GameTime(%f, %d) End with POS(%d ,%d), collisionPos: %d, BackDis(%d, %d), Speed(%d, %d), Rush(%d, %d)",
		float64(GameTime)/float64(TimeMili),
		GameTime,
		LeftGamer.GetGladiator().CurUnit, RightGamer.GetGladiator().CurUnit,
		collisionPos,
		LeftBack, RightBack,
		LeftGamer.GetGladiator().Speed, RightGamer.GetGladiator().Speed,
		LeftGamer.GetGladiator().Rush, RightGamer.GetGladiator().Rush,
	)
	return MyRoom.GetPackPlayerStates(), GameTime + KNOCK_BACK_TIME*TimeMili
}
