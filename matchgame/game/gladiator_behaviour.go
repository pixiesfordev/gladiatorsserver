package game

import (
	"fmt"
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/utility"
	"math"
	"time"

	log "github.com/sirupsen/logrus"
)

// 使用牌後更新手牌，使用的牌放回牌庫最後一張
func (g *Gladiator) UseSkill(skillID int) error {
	// log.Errorf("手牌: %v,   使用技能: %v   牌庫: %v", g.HandSkills, skillID, g.Deck)
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
	// log.Errorf("結果 手牌: %v,   牌庫: %v", g.HandSkills, g.Deck)
	return nil
}
func (g *Gladiator) SetRush(on bool) {
	if _, ok := g.Effects[gameJson.Enraged]; ok && !on {
		// 激怒中無法設定回非衝刺狀態
		return
	}
	g.IsRush = on
}

// ActiveSkill 啟用技能
func (g *Gladiator) ActiveSkill(jsonSkill gameJson.JsonSkill, on bool) error {
	// 消耗體力
	if int(g.CurVigor) < jsonSkill.Vigor {
		return fmt.Errorf("玩家%v 嘗試施放體力不足的技能(%v)", g.Owner.GetID(), jsonSkill.ID)
	}
	g.AddVigor(float64(-jsonSkill.Vigor)) // 消耗體力
	switch jsonSkill.Activation {
	case gameJson.Melee: // 肉搏技能
		if on {
			g.ActivedMeleeJsonSkill = &jsonSkill
		} else {
			g.ActivedMeleeJsonSkill = nil
		}
	case gameJson.Instant: // 即時技能
		if on {
			// 發動即時技能
			_, skill, err := g.createSkill(jsonSkill)
			if err != nil {
				return err
			}
			g.SpellInstantSkill(skill)
		}
	default:
		return fmt.Errorf("未定義的技能啟用類型: %v", jsonSkill.Activation)
	}
	return nil
}

// SpellInstantSkill 施放立即技能
func (myself *Gladiator) SpellInstantSkill(skill *Skill) {
	if skill.JsonSkill.Activation != gameJson.Instant {
		return
	}

	if skill.JsonSkill.Init != 0 {
		// 計算命中技能花費的時間
		dist := getDistBetweenGladiators()
		hitMiliSecs := int(math.Round((dist / skill.JsonSkill.Init) * 1000))
		// 延遲時間後命中目標
		time.AfterFunc(time.Duration(hitMiliSecs)*time.Millisecond, func() {
			myself.Spell(skill)
			knockback(myself.Opponent, myself, skill.JsonSkill.Knockback)
		})
	} else {
		myself.Spell(skill)
		knockback(myself.Opponent, myself, skill.JsonSkill.Knockback)
	}

}

// Spell 施放技能
func (myself *Gladiator) Spell(skill *Skill) {
	log.Infof("勇士 %s 發動技能: %v", myself.ID, skill.JsonSkill.ID)

	// 執行技能效果
	for _, effect := range skill.Effects {
		if effect == nil {
			log.Errorf("effect為nil")
			continue
		}
		// 如果施法者或目標為nil
		if effect.Target == nil || effect.Speller == nil {
			log.Errorf("target 或 speller 為nil, target: %v speller: %v", effect.Target, effect.Speller)
			break
		}
		// 如果施法者或目標死亡就跳過
		if !effect.Target.IsAlive || !effect.Speller.IsAlive {
			break
		}
		// 如果沒觸發成功就跳過
		if !utility.GetProbResult(effect.Prob) {
			continue
		}
		switch effect.Type {
		case gameJson.PDmg: // 物理攻擊
			dmg := AttackDmgModify(myself, effect.Target, effect)
			myself.Attack(effect.Target, dmg, effect.Type)
		case gameJson.MDmg: // 魔法攻擊
			dmg := AttackDmgModify(myself, effect.Target, effect)
			myself.Attack(effect.Target, dmg, effect.Type)
		case gameJson.TrueDmg: // 真實傷害攻擊
			dmg := 0.0
			if !effect.Target.ImmuneTo(TDMG) {
				dmg, _ = GetEffectValue[float64](effect, 0)
			}
			myself.Attack(effect.Target, int(dmg), effect.Type)
		case gameJson.RestoreHP: // 回復生命
			effect.Target.AddHp(effect.GetRestoreHPValue(), effect.Type, true)
		case gameJson.RestoreVigor: // 回復體力
			effect.Target.AddVigor(effect.GetRestoreVigorValue())
		case gameJson.Purge: // 移除負面狀態
			effect.Target.RemoveEffectsByTag(DEBUFF)
		default:
			effect.Target.AddEffect(effect)
		}
	}

	// 如果有ComboAttack就重複觸發技能
	if effects, ok := myself.Effects[gameJson.ComboAttack]; ok {
		if len(effects) != 0 {
			effects[0].AddDuration(-1)
			myself.Spell(skill)
		}
	}

}

func AttackDmgModify(myself, target *Gladiator, effect *Effect) int {
	dmg := 0.0
	if effect.Type == gameJson.PDmg && !effect.Target.ImmuneTo(PDMG) {
		multiple := myself.GetPDmgMultiplier()
		dmg, _ = GetEffectValue[float64](effect, 0)
		dmg = math.Round(dmg * multiple)
		dmg -= float64(effect.Target.GetPDef())
	} else if effect.Type == gameJson.MDmg && !effect.Target.ImmuneTo(MDMG) {
		multiple := myself.GetMDmgMultiplier()
		dmg, _ = GetEffectValue[float64](effect, 0)
		dmg = math.Round(dmg * multiple)
		dmg -= float64(effect.Target.GetMDef())
	}

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
			dmg = math.Round(float64(dmg) * (myself.CritDmg + extraCritDmg))
		}
		// 計算目標傷害調整百分比
		dmgMultiplier := 0.0
		for _, effects := range target.Effects {
			for _, v := range effects {
				dmgMultiplier += v.GetTakeDmgMultiple()
			}
		}
		dmg = float64(dmg) * (1 + dmgMultiplier)
	}
	return int(math.Round(dmg))
}

// Attack 造成傷害
func (myself *Gladiator) Attack(target *Gladiator, dmg int, effectType gameJson.EffectType) {
	// 造成傷害
	target.AddHp(-dmg, effectType, true)
	// 觸發攻擊後效果
	target.TriggerBuffer_AfterBeAttack(dmg)
	myself.TriggerBuffer_AfterAttack(dmg)

}

func (g *Gladiator) Move() {
	// 無法移動就return
	if g.ImmuneTo(MOVE) {
		return
	}

	// 計算移動向量
	moveVec := g.FaceDir.Multiply(g.GetSpd() * TickTimePass)
	g.CurPos = g.CurPos.Add(moveVec)
	// 限制在戰場範圍內
	if g.CurPos.X > BATTLEFIELD.X {
		g.CurPos.X = BATTLEFIELD.X
	} else if g.CurPos.X < -BATTLEFIELD.X {
		g.CurPos.X = -BATTLEFIELD.X
	}

	if g.CurPos.Y > BATTLEFIELD.Y {
		g.CurPos.Y = BATTLEFIELD.Y
	} else if g.CurPos.Y < -BATTLEFIELD.Y {
		g.CurPos.Y = -BATTLEFIELD.Y
	}
}
