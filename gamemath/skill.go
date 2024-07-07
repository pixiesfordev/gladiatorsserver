package main

import (
	gameJson "gladiatorsGoModule/gamejson"
)

type Skill struct {
	JsonSkill gameJson.JsonSkill
	Speller   *Gladiator
	Init      int
	Vigor     int
	Knockback int
	Effects   []Effect
}

func NewSkill(speller *Gladiator, opponent *Gladiator, jsonSkill gameJson.JsonSkill) (Skill, error) {

	skill := Skill{
		JsonSkill: jsonSkill,
		Speller:   speller,
		Vigor:     jsonSkill.Vigor,
		Init:      jsonSkill.Initiative,
		Knockback: jsonSkill.Knockback,
	}

	effects := make([]Effect, 0)

	if jsonSkillEffects, ok := gameJson.SkillEffectDataDic[jsonSkill.ID]; ok {
		for _, jsonSkillEffect := range jsonSkillEffects {
			var skillEffectTarget *Gladiator
			if jsonSkillEffect.Target == "Myslef" {
				skillEffectTarget = speller
			} else if jsonSkillEffect.Target == "Enemy" {
				skillEffectTarget = opponent
			}
			for _, v := range jsonSkillEffect.Effects {
				effect := Effect{
					Type:    v.Type,
					Value:   v.Value,
					Speller: speller,
					Target:  skillEffectTarget,
					Prob:    v.Prob,
				}
				effects = append(effects, effect)
			}
		}
	}
	skill.Effects = effects
	return skill, nil
}
