package main

import (
	log "github.com/sirupsen/logrus"
	utility "gladiatorsGoModule/utility"
	"math/rand"
	"time"
)

var Rnd *rand.Rand

func main() {
	log.Infof("======gamemath開始測試======")

	source := rand.NewSource(time.Now().UnixNano())
	Rnd = rand.New(source)

}

func Melee(g1 *Gladiator, g2 *Gladiator, g1Skill *Skill, g2Skill *Skill) {

	// 計算雙方先攻值
	g1SpellInit := g1.GetInit() + g1Skill.Init
	g2SpellInit := g2.GetInit() + g2Skill.Init

	// 雙方技能施放
	if g1SpellInit > g2SpellInit { // g1先攻
		g1.Spell(g1Skill)
		g2.Spell(g2Skill)
	} else if g1SpellInit < g2SpellInit { // g2先攻
		g2.Spell(g2Skill)
		g1.Spell(g1Skill)
	} else { // 先攻值一樣的話就隨機一方先攻
		if utility.GetProbResult(0.5) {
			g1.Spell(g1Skill)
			g2.Spell(g2Skill)
		} else {
			g2.Spell(g2Skill)
			g1.Spell(g1Skill)
		}
	}

	// 雙方擊退
	g1Knockback := g1.GetKnockback() + g1Skill.Knockback
	g2Knockback := g2.GetKnockback() + g2Skill.Knockback

	g1.DoKnockback(float64(g2Knockback))
	g2.DoKnockback(float64(g1Knockback))
}
