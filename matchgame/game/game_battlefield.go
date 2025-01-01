package game

import (
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/utility"
	"matchgame/packet"
	"time"

	log "github.com/sirupsen/logrus"
)

var BATTLEFIELD utility.Vector2 = utility.Vector2{X: 20, Y: 20}        // 戰鬥場地
var KnockbackAngleRange int = 15                                       // 擊退角度範圍
var GLADIATOR_POS_LEFT utility.Vector2 = utility.Vector2{X: -16, Y: 0} // 左方角鬥士初始座標
var GLADIATOR_POS_RIGHT utility.Vector2 = utility.Vector2{X: 16, Y: 0} // 右方角鬥士初始座標

func knockback(targetGladiator, opponentGladiator *Gladiator, knockbackDist float64) {
	if targetGladiator.ImmuneTo(KNOCKBACK) || knockbackDist <= 0 {
		return
	}
	beforePos := targetGladiator.CurPos

	// 產生一個隨機的角度變化
	angleChange, err := utility.RandomFloatBetweenInts(-KnockbackAngleRange, KnockbackAngleRange)
	if err != nil {
		log.Errorf("knockback錯誤 %v", err)
		return
	}

	// 將角度變化轉換為弧度
	radians := utility.DegreesToRadians(angleChange)
	// 計算新的方向向量
	newDir := targetGladiator.FaceDir.Multiply(-1).RotateVector(radians)
	// 根據新方向和擊退距離計算新位置
	targetGladiator.CurPos = targetGladiator.CurPos.Add(newDir.Normalize().Multiply(knockbackDist))

	isKnockWall := false
	// 檢查是否有撞牆
	if targetGladiator.CurPos.X <= -BATTLEFIELD.X {
		targetGladiator.CurPos.X = -BATTLEFIELD.X
		targetGladiator.knockWall()
		isKnockWall = true
	} else if targetGladiator.CurPos.X >= BATTLEFIELD.X {
		targetGladiator.CurPos.X = BATTLEFIELD.X
		targetGladiator.knockWall()
		isKnockWall = true
	}

	if targetGladiator.CurPos.Y <= -BATTLEFIELD.Y {
		targetGladiator.CurPos.Y = -BATTLEFIELD.Y
		targetGladiator.knockWall()
		isKnockWall = true
	} else if targetGladiator.CurPos.Y >= BATTLEFIELD.Y {
		targetGladiator.CurPos.Y = BATTLEFIELD.Y
		targetGladiator.knockWall()
		isKnockWall = true
	}

	updateGladiatorsFaceDir(targetGladiator, opponentGladiator)

	pack := packet.Pack{
		CMD: packet.KNOCKBACK_TOCLIENT,
		Content: &packet.Knockback_ToClient{
			PlayerID:      targetGladiator.Owner.GetID(),
			BeforePos:     beforePos.Round2(),
			KnockbackDist: knockbackDist,
			AfterPos:      targetGladiator.CurPos.Round2(),
			KnockWall:     isKnockWall,
		},
	}
	MyRoom.BroadCastPacket(-1, pack)
}

// 更新角鬥士的面朝方向
func updateGladiatorsFaceDir(g1, g2 *Gladiator) {
	directionToG2 := g2.CurPos.Sub(g1.CurPos).Normalize()
	directionToG1 := g1.CurPos.Sub(g2.CurPos).Normalize()
	g1.FaceDir = directionToG2
	g2.FaceDir = directionToG1
}

// 撞牆
func (myself *Gladiator) knockWall() {
	time.AfterFunc(time.Duration(Knockwall_DmgDelayMiliSecs)*time.Millisecond, func() {
		myself.AddHp(-Knockwall_Dmg, gameJson.PDmg, true)
	})
}
