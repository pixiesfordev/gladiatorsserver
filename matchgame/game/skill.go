package game

import (
	"gladiatorsGoModule/gameJson"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type Skill struct {
	JsonSkill gameJson.JsonSkill
	Speller   *Gladiator
	Effects   []*Effect
}

func NewSkill(speller *Gladiator, opponent *Gladiator, jsonSkill gameJson.JsonSkill) (*Skill, error) {

	skill := Skill{
		JsonSkill: jsonSkill,
		Speller:   speller,
	}

	effects := make([]*Effect, 0)

	if jsonSkillEffects, ok := gameJson.SkillEffectDataDic[jsonSkill.ID]; ok {
		for _, jsonSkillEffect := range jsonSkillEffects {
			var skillEffectTarget *Gladiator
			if jsonSkillEffect.Target == "Myself" {
				skillEffectTarget = speller
			} else if jsonSkillEffect.Target == "Enemy" {
				skillEffectTarget = opponent
			} else {
				log.Errorf("jsonSkillEffect.Target錯誤: %v", jsonSkillEffect.Target)
				continue
			}
			for _, v := range jsonSkillEffect.Effects {
				effect, err := NewEffect(v.Type, v.Value, speller, skillEffectTarget, v.Prob, false)
				if err != nil {
					return nil, err
				}
				effects = append(effects, effect)
			}
		}
	}
	skill.Effects = effects
	return &skill, nil
}

// NewBaseKnockSkill 創建基礎擊退技能(肉搏但沒有施放肉搏技能時會使用基本擊退技能)
func NewBaseKnockSkill(speller *Gladiator, opponent *Gladiator) (*Skill, error) {
	skill := Skill{
		Speller: speller,
	}
	effects := make([]*Effect, 0)
	effect, err := NewEffect(gameJson.PDmg, strconv.Itoa(speller.GetStr()), speller, opponent, 1, false)
	if err != nil {
		return nil, err
	}
	effects = append(effects, effect)
	skill.Effects = effects
	return &skill, nil
}
