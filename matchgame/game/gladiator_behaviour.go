package game

import (
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/utility"
	"matchgame/packet"
	"math"

	log "github.com/sirupsen/logrus"
)

// 使用牌後更新手牌，使用的牌放回牌庫最後一張
func (g *Gladiator) UseSkill(skillID int) bool {
	_, useSkillIdx, err := g.GetSkill(skillID)
	if err != nil {
		return false
	}
	// 將使用的牌放到牌庫最底下
	g.Deck = append(g.Deck, g.HandSkills[useSkillIdx])

	// 將手牌第4張牌(下一張牌)補到使用的牌位置
	g.HandSkills[useSkillIdx] = g.HandSkills[3]

	// 從牌庫中按順序抽出一張作為新的第4張牌
	g.HandSkills[3] = g.Deck[0] // 將牌庫頂部的牌補到第4張手牌中(下一張牌)
	g.Deck = g.Deck[1:]         // 從剛抽出的牌移除

	return true
}

func (g *Gladiator) SetRush(on bool) {
	g.IsRush = on
	// 衝刺狀態更改要送封包回Client
	pack := packet.Pack{
		CMD:    packet.PLAYERACTION_TOCLIENT,
		PackID: -1,
		Content: packet.PlayerAction_ToClient{
			PlayerDBID: g.OwnerID,
			ActionType: packet.Action_Rush,
			ActionContent: packet.PackAction_Rush_ToClient{
				On: on,
			},
		},
	}
	MyRoom.BroadCastPacket(-1, pack)
}

// ActiveSkill 啟用技能
func (g *Gladiator) ActiveSkill(jsonSkill gameJson.JsonSkill, on bool) {
	switch jsonSkill.Activation {
	case "Melee": // 肉搏技能
		if on {
			g.ActivedMeleeJsonSkill = &jsonSkill
		} else {
			g.ActivedMeleeJsonSkill = nil
		}
	case "Instant": // 即時技能
		if on {
			// 發動即時技能
		}
	default:
		log.Errorf("未定義的技能啟用類型: %v", jsonSkill.Activation)
	}
}

// 執行擊退
func (myself *Gladiator) DoKnockback(knockbackValue float64) {

	//對擊退免疫就返回
	immuneTypes := myself.GetImmuneTypes()
	if _, ok := immuneTypes[Immune_Knockback]; ok {
		return
	}

	// 執行擊退位移
	if myself.LeftSide {
		myself.CurPos -= knockbackValue
	} else {
		myself.CurPos += knockbackValue
	}

	// 檢查是否有撞牆
	if myself.LeftSide && myself.CurPos <= -WallPos {
		myself.CurPos = -WallPos
		myself.knockWall()
	} else if !myself.LeftSide && myself.CurPos >= WallPos {
		myself.CurPos = WallPos
		myself.knockWall()
	}

}

// 撞牆
func (myself *Gladiator) knockWall() {
}

// Spell 發動技能
func (myself *Gladiator) Spell(skill *Skill) {
	if myself == nil || skill == nil {
		return
	}

	// 檢查是否通過技能施法條件
	if !myself.IsPassingSpellCondition(skill) {
		return
	}

	// 執行技能效果
	for _, effect := range skill.Effects {

		// 如果施法者或目標死亡就跳過
		if effect.Target.IsAlive() || !effect.Speller.IsAlive() {
			continue
		}

		// 如果沒觸發成功就跳過
		if !utility.GetProbResult(effect.Prob) {
			continue
		}

		switch effect.Type {
		case gameJson.PDmg: // 物理攻擊
			multiple := (1 + effect.GetPDmgMultiple())
			effectDmg := float64(effect.GetPDmgValue())
			str := float64(myself.GetStr())
			dmg := int(math.Round(str * effectDmg * multiple))
			myself.Attack(effect.Target, dmg, effect.Target.GetPDef())
		case gameJson.MDmg: // 魔法攻擊
			multiple := (1 + effect.GetMDmgMultiple())
			effectDmg := float64(effect.GetMDmgValue())
			str := float64(myself.GetStr())
			dmg := int(math.Round(str * effectDmg * multiple))
			myself.Attack(effect.Target, dmg, effect.Target.GetMDef())
		case gameJson.RestoreHP: // 回復生命
			effect.Target.AddHp(effect.GetRestoreHPValue())
		case gameJson.RestoreVigor: // 回復體力
			effect.Target.AddVigor(effect.GetRestoreVigorValue())
		case gameJson.Purge: // 移除負面狀態
			effect.Target.RemoveEffectsByBufferType(Debuff)
		default:
			if effect.IsBuffer() { // 賦予狀態
				myself.AddEffect(&effect)
			}
		}
	}

	log.Infof("勇士 %s 發動技能: %v", myself.ID, skill.JsonSkill.ID)

	// 如果有ComboAttack就重複觸發技能
	if effects, ok := myself.Effects[gameJson.ComboAttack]; ok {
		if len(effects) != 0 {
			effects[0].AddDuration(-1)
			myself.Spell(skill)
		}
	}

}

// IsPassingSpellCondition 施法條件檢查, 全部偷過才返回true(代表可以觸發)
func (myself *Gladiator) IsPassingSpellCondition(skill *Skill) bool {
	if !myself.IsAlive() {
		return false
	}
	for _, effects := range myself.Effects {
		for _, v := range effects {
			switch v.Type {
			case gameJson.Condition_SkillVigorBelow:
				value, err := GetEffectValue[int](v, 0)
				if err != nil {
					log.Errorf("%v錯誤: %v", v.Type, err)
					return false
				}
				if skill.JsonSkill.Vigor > value {
					return false
				}
			}
		}
	}
	return true
}

// Attack 造成傷害
func (myself *Gladiator) Attack(target *Gladiator, dmg int, def int) {
	dealDmg := dmg - def

	// 計算目標狀態減傷
	multiple := 0.0
	for _, effects := range target.Effects {
		for _, v := range effects {
			multiple += v.GetTakeDmgMultiple()
		}
	}
	dealDmg = int(math.Round(float64(dealDmg) * (1 + multiple)))

	// 計算爆擊
	crit := myself.GetCrit()
	extraCritDmg := 0.0
	if crit > 1 { // 溢出的爆擊率要加到爆擊傷害上
		extraCritDmg = 1 - crit
	}
	if utility.GetProbResult(myself.GetCrit()) {
		dealDmg = int(math.Round(float64(dealDmg) * (myself.CritDmg + extraCritDmg)))
	}

	target.TriggerBuffer_AfterBeAttack(dealDmg)
	myself.TriggerBuffer_AfterAttack(dealDmg)

	target.AddHp(dealDmg)
}

func (g *Gladiator) Move() {
	// 無法移動就return
	if !g.CanMove() {
		return
	}

	movePos := g.GetSpd() * g.GetDir() * TickTimePass
	// log.Infof("g.CurPos1 %v g.GetSpd() %v g.GetDir() %v  TickTimePass %v movePos %v ", g.CurPos, g.GetSpd(), g.GetDir(), TickTimePass, movePos)
	g.CurPos += movePos
	if g.CurPos > WallPos {
		g.CurPos = WallPos
	} else if g.CurPos <= -WallPos {
		g.CurPos = -WallPos
	}
}
