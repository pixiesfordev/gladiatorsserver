package game

import (
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/utility"

	"matchgame/packet"
)

// melee 雙方進行肉搏
func melee() {
	if MyRoom.Gamers[0] == nil && MyRoom.Gamers[0].GetGladiator() == nil && MyRoom.Gamers[1] == nil && MyRoom.Gamers[1].GetGladiator() == nil {
		return
	}
	g1 := MyRoom.Gamers[0].GetGladiator()
	g2 := MyRoom.Gamers[1].GetGladiator()

	if !g1.IsAlive() || !g2.IsAlive() {
		return
	}

	// <<<<<<<<<<初始化雙方肉搏技能>>>>>>>>>>
	g1SpellInit := 0.0
	var g1Skill *Skill
	g2SpellInit := 0.0
	var g2Skill *Skill
	if g1.ActivedMeleeJsonSkill != nil {
		g1SpellInit, g1Skill, _ = g1.createSkill(*g1.ActivedMeleeJsonSkill)
		g1.ActivedMeleeJsonSkill = nil
	}
	if g2.ActivedMeleeJsonSkill != nil {
		g2SpellInit, g2Skill, _ = g2.createSkill(*g2.ActivedMeleeJsonSkill)
		g2.ActivedMeleeJsonSkill = nil
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
	g1AttackPos := g1.CurPos
	g1SkillKnockback := 0.0
	if g1Skill != nil {
		g1SkillKnockback += g1Skill.JsonSkill.Knockback
	}
	g1Knockback := g1.GetKnockback() + g1SkillKnockback
	if g2.ImmuneTo(KNOCKBACK) {
		g1Knockback = 0
	}
	g2.DoKnockback(g1Knockback)

	g2AttackPos := g2.CurPos
	g2SkillKnockback := 0.0
	if g2Skill != nil {
		g2SkillKnockback += g2Skill.JsonSkill.Knockback
	}
	g2Knockback := g2.GetKnockback() + g2SkillKnockback + 5
	if g1.ImmuneTo(KNOCKBACK) {
		g2Knockback = 0
	}
	g1.DoKnockback(g2Knockback)

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
	if player, ok := MyRoom.Gamers[0].(*Player); ok {
		packMelee := packet.Pack{
			CMD: packet.MELEE_TOCLIENT,
			Content: &packet.Melee_ToClient{
				MyAttack: packet.PackMelee{
					Knockback:   g1Knockback,
					SkillID:     g1SkillID,
					MeleePos:    g1AttackPos,
					CurPos:      g1.CurPos,
					EffectTypes: g1.GetEffectStrs(),
				},
				OpponentAttack: packet.PackMelee{
					Knockback:   g2Knockback,
					SkillID:     g2SkillID,
					MeleePos:    g2AttackPos,
					CurPos:      g2.CurPos,
					EffectTypes: g2.GetEffectStrs(),
				},
				MyHandSkillIDs: g1.GetHandSkills(),
			},
		}
		player.SendPacketToPlayer(packMelee)
	}
	if player, ok := MyRoom.Gamers[1].(*Player); ok {
		packMelee := packet.Pack{
			CMD: packet.MELEE_TOCLIENT,
			Content: &packet.Melee_ToClient{
				MyAttack: packet.PackMelee{
					Knockback:   g2Knockback,
					SkillID:     g2SkillID,
					CurPos:      g2.CurPos,
					EffectTypes: g2.GetEffectStrs(),
				},
				OpponentAttack: packet.PackMelee{
					Knockback:   g1Knockback,
					SkillID:     g1SkillID,
					CurPos:      g1.CurPos,
					EffectTypes: g1.GetEffectStrs(),
				},
				MyHandSkillIDs: g2.GetHandSkills(),
			},
		}
		player.SendPacketToPlayer(packMelee)
	}

}

// 雙方執行技能
func bothCastSpell(first, second *Gladiator, firstSkill, secondSkill *Skill) {
	if firstSkill != nil {
		second.SkillHit(firstSkill)
	}
	if secondSkill != nil {
		first.SkillHit(secondSkill)
	}
}
