package game

import (
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/utility"
	"math"
	"time"

	log "github.com/sirupsen/logrus"
)

// 使用牌後更新手牌，使用的牌放回牌庫最後一張
func (g *Gladiator) UseSkill(skillID int) error {
	_, useSkillIdx, err := g.GetSkill(skillID)
	if err != nil {
		return err
	}
	// 將使用的牌放到牌庫最底下
	g.Deck = append(g.Deck, g.HandSkills[useSkillIdx])
	// 將手牌第4張牌(下一張牌)補到使用的牌位置
	g.HandSkills[useSkillIdx] = g.HandSkills[3]
	// 從牌庫中按順序抽出一張作為新的第4張牌
	g.HandSkills[3] = g.Deck[0] // 將牌庫頂部的牌補到第4張手牌中(下一張牌)
	g.Deck = g.Deck[1:]         // 從剛抽出的牌移除

	return nil
}

func (g *Gladiator) SetRush(on bool) {
	g.IsRush = on
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
			_, skill, err := g.createSkill(jsonSkill)
			if err != nil {
				return
			}
			g.SpellConstSkill(skill)
		}
	default:
		log.Errorf("未定義的技能啟用類型: %v", jsonSkill.Activation)
	}
}

// 執行擊退
func (myself *Gladiator) DoKnockback(knockbackValue float64) {

	if knockbackValue <= 0 {
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
	time.AfterFunc(time.Duration(Knockwall_DmgDelayMiliSecs)*time.Millisecond, func() {
		myself.AddHp(-Knockwall_Dmg, true)
	})
}

// SpellConstSkill 施放立即技能
func (myself *Gladiator) SpellConstSkill(skill *Skill) {
	if skill.JsonSkill.Activation != "Instant" {
		return
	}
	// 計算命中技能花費的時間
	hitMiliSecs := 0
	if skill.JsonSkill.Init != 0 {
		dist := getDistBetweenGladiators()
		hitMiliSecs = int(math.Round((dist / skill.JsonSkill.Init) * 1000))
	}
	// 延遲時間後命中目標
	time.AfterFunc(time.Duration(hitMiliSecs)*time.Millisecond, func() {
		myself.SkillHit(skill)
	})
}

// SkillHit 技能命中目標
func (myself *Gladiator) SkillHit(skill *Skill) {
	if skill == nil {
		return
	}
	// 執行技能效果
	for _, effect := range skill.Effects {
		// 如果施法者或目標死亡就跳過
		if !effect.Target.IsAlive() || !effect.Speller.IsAlive() {
			continue
		}
		// 如果沒觸發成功就跳過
		if !utility.GetProbResult(effect.Prob) {
			continue
		}
		switch effect.Type {
		case gameJson.PDmg: // 物理攻擊
			dmg := 0.0
			if !myself.ImmuneTo(PDMG) {
				multiple := myself.GetPDmgMultiplier()
				dmg, _ := GetEffectValue[float64](effect, 0)
				dmg = math.Round(dmg * multiple)
				dmg -= float64(effect.Target.GetPDef())
			}
			myself.Attack(effect.Target, int(dmg))
		case gameJson.MDmg: // 魔法攻擊
			dmg := 0.0
			if !myself.ImmuneTo(MDMG) {
				multiple := myself.GetMDmgMultiplier()
				dmg, _ := GetEffectValue[float64](effect, 0)
				dmg = math.Round(dmg * multiple)
				dmg -= float64(effect.Target.GetMDef())
			}
			myself.Attack(effect.Target, int(dmg))
		case gameJson.RestoreHP: // 回復生命
			effect.Target.AddHp(effect.GetRestoreHPValue(), true)
		case gameJson.RestoreVigor: // 回復體力
			effect.Target.AddVigor(effect.GetRestoreVigorValue())
		case gameJson.Purge: // 移除負面狀態
			effect.Target.RemoveEffectsByTag(DEBUFF)
		default:
			myself.AddEffect(effect)
		}
	}

	log.Infof("勇士 %s 發動技能: %v", myself.ID, skill.JsonSkill.ID)

	// 如果有ComboAttack就重複觸發技能
	if effects, ok := myself.Effects[gameJson.ComboAttack]; ok {
		if len(effects) != 0 {
			effects[0].AddDuration(-1)
			myself.SkillHit(skill)
		}
	}

}

// Attack 造成傷害
func (myself *Gladiator) Attack(target *Gladiator, dmg int) {
	// 傷害不會小於0
	if dmg <= 0 {
		dmg = 0
	} else {
		// 計算爆擊
		crit := myself.GetCrit()
		extraCritDmg := 0.0
		if crit > 1 { // 溢出的爆擊率要加到爆擊傷害上
			extraCritDmg = 1 - crit
		}
		if utility.GetProbResult(myself.GetCrit()) {
			dmg = int(math.Round(float64(dmg) * (myself.CritDmg + extraCritDmg)))
		}
		// 計算目標傷害調整百分比
		dmgMltiple := 0.0
		for _, effects := range target.Effects {
			for _, v := range effects {
				dmgMltiple += v.GetTakeDmgMultiple()
			}
		}
		dmg = int(math.Round(float64(dmg) * (1 + dmgMltiple)))
	}
	// 造成傷害
	target.AddHp(-dmg, true)
	// 觸發攻擊後效果
	target.TriggerBuffer_AfterBeAttack(dmg)
	myself.TriggerBuffer_AfterAttack(dmg)

}

func (g *Gladiator) Move() {
	// 無法移動就return
	if g.ImmuneTo(MOVE) {
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
