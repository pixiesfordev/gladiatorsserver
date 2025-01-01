package game

import (
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/utility"

	"matchgame/packet"

	log "github.com/sirupsen/logrus"
)

// melee 雙方進行肉搏
func melee(gamer1, gamer2 Gamer, g1, g2 *Gladiator) {

	if !g1.IsAlive || !g2.IsAlive {
		log.Infof("取消melee，一方角鬥已死亡 g1: %v   g2: %v", g1.IsAlive, g2.IsAlive)
		return
	}

	// <<<<<<<<<<初始化雙方肉搏技能>>>>>>>>>>
	g1SpellInit := g1.GetInit() // 先攻值
	var g1Skill *Skill
	g2SpellInit := g2.GetInit() // 先攻值
	var g2Skill *Skill
	var err error
	if g1.ActivedMeleeJsonSkill != nil {
		// if g1.ActivedMeleeJsonSkill.Vigor <= int(g1.CurVigor) {
		g1SpellInit, g1Skill, err = g1.createSkill(*g1.ActivedMeleeJsonSkill)
		if err != nil {
			log.Errorf("createSkill錯誤: %v", err)
		}
		g1.ActivedMeleeJsonSkill = nil
	}
	// 如果沒有施放肉搏技能，則施放基礎擊退技能
	if g1Skill == nil {
		g1SpellInit, g1Skill, err = g1.createBaseKnockSkill()
		if err != nil {
			log.Errorf("createBaseKnockSkill錯誤: %v", err)
		}
	}
	if g2.ActivedMeleeJsonSkill != nil {
		// if g2.ActivedMeleeJsonSkill.Vigor <= int(g1.CurVigor) {
		g2SpellInit, g2Skill, err = g2.createSkill(*g2.ActivedMeleeJsonSkill)
		if err != nil {
			log.Errorf("createSkill錯誤: %v", err)
		}
		g2.ActivedMeleeJsonSkill = nil
		// } else {
		// 	log.Errorf("%v 體力不足無法施放技能 %v", gamer2.GetID(), g2.ActivedMeleeJsonSkill.ID)
		// }
	}
	// 如果沒有施放肉搏技能，則施放基礎擊退技能
	if g2Skill == nil {
		g2SpellInit, g2Skill, err = g2.createBaseKnockSkill()
		if err != nil {
			log.Errorf("createBaseKnockSkill錯誤: %v", err)
		}
	}

	if g1SpellInit > g2SpellInit {
		bothCastSpell(g1, g2, g1Skill, g2Skill) // g1先攻
	} else if g1SpellInit < g2SpellInit {
		bothCastSpell(g2, g1, g2Skill, g1Skill) // g2先攻
	} else { // 先攻值一樣就隨機一方先攻
		if utility.GetProbResult(0.5) {
			bothCastSpell(g1, g2, g1Skill, g2Skill) // 隨機g1先攻
		} else {
			bothCastSpell(g2, g1, g2Skill, g1Skill) // 隨機g2先攻
		}
	}

	// <<<<<<<<<<擊退>>>>>>>>>>
	g1SkillKnockback := 0.0
	if g1Skill != nil {
		g1SkillKnockback += g1Skill.JsonSkill.Knockback
	}
	g1Knockback := g1.GetKnockback() + g1SkillKnockback
	knockback(g2, g1, g1Knockback)

	g2SkillKnockback := 0.0
	if g2Skill != nil {
		g2SkillKnockback += g2Skill.JsonSkill.Knockback
	}
	g2Knockback := g2.GetKnockback() + g2SkillKnockback
	knockback(g1, g2, g2Knockback)

	// 增加暈眩狀態
	g1KnockDizzy, err := NewEffect(gameJson.Dizzy, "2", g2, g1, 1, false)
	if err == nil {
		g1.AddEffect(g1KnockDizzy)
	}
	g2KnockDizzy, err := NewEffect(gameJson.Dizzy, "2", g1, g2, 1, false)
	if err == nil {
		g2.AddEffect(g2KnockDizzy)
	}

	// <<<<<<<<<<Melee封包給Client>>>>>>>>>>
	g1SkillID := 0
	g2SkillID := 0
	if g1Skill != nil {
		g1SkillID = g1Skill.JsonSkill.ID
	}
	if g2Skill != nil {
		g2SkillID = g2Skill.JsonSkill.ID
	}
	g1SkillOnID := 0
	g2SkillOnID := 0
	if g1.ActivedMeleeJsonSkill != nil {
		g1SkillOnID = g1.ActivedMeleeJsonSkill.ID
	}
	if g2.ActivedMeleeJsonSkill != nil {
		g2SkillOnID = g2.ActivedMeleeJsonSkill.ID
	}

	if p1, ok := gamer1.(*Player); ok {
		hands := g1.GetHandSkills()
		newSkillID := 0
		if g1Skill != nil {
			newSkillID = hands[3]
		}

		packMelee := packet.Pack{
			CMD: packet.MELEE_TOCLIENT,
			Content: &packet.Melee_ToClient{
				MyAttack: packet.PackMelee{
					SkillID:     g1SkillID,
					EffectDatas: g1.GetPackEffects(),
				},
				OpponentAttack: packet.PackMelee{
					SkillID:     g2SkillID,
					EffectDatas: g2.GetPackEffects(),
				},
				NewSkilID:  newSkillID,
				SkillOnID:  g1SkillOnID,
				HandSkills: hands,
			},
		}
		p1.SendPacketToPlayer(packMelee)

	}

	if p2, ok := gamer2.(*Player); ok {
		hands := g2.GetHandSkills()
		newSkillID := 0
		if g2Skill != nil {
			newSkillID = hands[3]
		}
		packMelee := packet.Pack{
			CMD: packet.MELEE_TOCLIENT,
			Content: &packet.Melee_ToClient{
				MyAttack: packet.PackMelee{
					SkillID:     g2SkillID,
					EffectDatas: g2.GetPackEffects(),
				},
				OpponentAttack: packet.PackMelee{
					SkillID:     g1SkillID,
					EffectDatas: g1.GetPackEffects(),
				},
				NewSkilID:  newSkillID,
				SkillOnID:  g2SkillOnID,
				HandSkills: hands,
			},
		}
		p2.SendPacketToPlayer(packMelee)
	}

}

// 雙方執行技能
func bothCastSpell(g1, g2 *Gladiator, g1Skill, g2Skill *Skill) {
	if g1Skill != nil {
		g1.Spell(g1Skill)
	}
	if g2Skill != nil {
		g2.Spell(g2Skill)
	}
}
