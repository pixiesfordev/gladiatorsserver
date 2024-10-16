package game

import (
	"gladiatorsGoModule/gameJson"
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
			if jsonSkillEffect.Target == "Myslef" {
				skillEffectTarget = speller
			} else if jsonSkillEffect.Target == "Enemy" {
				skillEffectTarget = opponent
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
