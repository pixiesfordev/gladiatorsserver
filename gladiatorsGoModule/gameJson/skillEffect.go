package gameJson

import (
	"encoding/json"
	"fmt"
)

type JsonSkillEffect struct {
	ID           string `json:"ID"`
	SkillID      int    `json:"SkillID"`
	Target       string `json:"Target"`
	EnemyEffects []SkillEffect
	MyselEffects []SkillEffect
}

type SkillEffect struct {
	Type  string
	Value int
	Prob  float64
}

var SkillEffectDataDic = make(map[int][]JsonSkillEffect)

func (jsonData JsonSkillEffect) UnmarshalJSONData(jsonName string, jsonBytes []byte) (map[interface{}]interface{}, error) {
	var wrapper map[string][]map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &wrapper); err != nil {
		return nil, err
	}

	datas, ok := wrapper[jsonName]
	if !ok {
		return nil, fmt.Errorf("找不到key值: %s", jsonName)
	}

	items := make(map[interface{}]interface{})
	for _, rawEffect := range datas {
		json := JsonSkillEffect{}
		json.ID = rawEffect["ID"].(string)
		json.SkillID = int(rawEffect["SkillID"].(float64))
		json.Target = rawEffect["Target"].(string)

		// 處理Effect
		if json.Target == "Enemy" {
			json.EnemyEffects = parseEffects(rawEffect)
		} else if json.Target == "Myself" {
			json.MyselEffects = parseEffects(rawEffect)
		}

		items[json.ID] = json
		AddToSkillEffectDataDic(json)
	}

	return items, nil
}

func parseEffects(data map[string]interface{}) []SkillEffect {
	effects := []SkillEffect{}
	for i := 1; ; i++ {
		typeKey := fmt.Sprintf("EffectType%d", i)
		valueKey := fmt.Sprintf("EffectValue%d", i)
		probKey := fmt.Sprintf("EffectProb%d", i)

		effectType, typeOk := data[typeKey]
		if !typeOk {
			break
		}

		effectValue := int(data[valueKey].(float64))
		effectProb := data[probKey].(float64)

		effects = append(effects, SkillEffect{
			Type:  effectType.(string),
			Value: effectValue,
			Prob:  effectProb,
		})
	}
	return effects
}

func AddToSkillEffectDataDic(jsonSkillEffect JsonSkillEffect) {
	if len(SkillEffectDataDic[jsonSkillEffect.SkillID]) != 0 {
		SkillEffectDataDic[jsonSkillEffect.SkillID] = append(SkillEffectDataDic[jsonSkillEffect.SkillID], jsonSkillEffect)
	} else {
		SkillEffectDataDic[jsonSkillEffect.SkillID] = []JsonSkillEffect{jsonSkillEffect}
	}
}
