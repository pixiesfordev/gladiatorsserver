package main

import (
	gameJson "gladiatorsGoModule/gamejson"
)

type Skill struct {
	JsonSkill gameJson.JsonSkill
	Speller   *Gladiator
	Target    *Gladiator
	Vigor     int
	Effects   []Effect
}

func NewSkill(speller *Gladiator, target *Gladiator, jsonSkill gameJson.JsonSkill) (Skill, error) {

	skill := Skill{
		JsonSkill: jsonSkill,
		Speller:   speller,
		Target:    target,
		Vigor:     jsonSkill.Vigor,
	}

	effects := make([]Effect, 0)

	if jsonSkillEffects, ok := gameJson.SkillEffectDataDic[jsonSkill.ID]; ok {
		for _, jsonSkillEffect := range jsonSkillEffects {
			var skillEffectTarget *Gladiator
			if jsonSkillEffect.Target == "Myslef" {
				skillEffectTarget = target
			} else if jsonSkillEffect.Target == "Enemy" {
				skillEffectTarget = speller
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
