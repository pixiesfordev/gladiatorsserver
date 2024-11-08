package gameJson

import (
	"encoding/json"
	"fmt"
	"gladiatorsGoModule/logger"
	"gladiatorsGoModule/utility"

	log "github.com/sirupsen/logrus"
)

type JsonSkill struct {
	ID         int            `json:"ID"`
	Activation ActivationType `json:"Activation"`
	Cost       int            `json:"Cost"`
	Init       float64        `json:"Init"`
	Vigor      int            `json:"Vigor"`
	Knockback  float64        `json:"Knockback"`
	Type       SkillType      `json:"Type"`
	Divine     string         `json:"Divine"`
}

// SkillType 技能類型
type SkillType string

const (
	TALENT SkillType = "Talent" // 天賦技能
	NORMAL SkillType = "Normal" //  一般技能
	DIVINE SkillType = "Divine" // 神祉技能
)

// ActivationType 技能類型
type ActivationType string

const (
	Melee   ActivationType = "Melee"   // 肉搏
	Instant ActivationType = "Instant" //  立即技能
)

func (jsonData JsonSkill) UnmarshalJSONData(jsonName string, jsonBytes []byte) (map[interface{}]interface{}, error) {
	var wrapper map[string][]JsonSkill
	if err := json.Unmarshal(jsonBytes, &wrapper); err != nil {
		return nil, err
	}

	datas, ok := wrapper[jsonName]
	if !ok {
		return nil, fmt.Errorf("找不到key值: %s", jsonName)
	}

	items := make(map[interface{}]interface{})
	for _, item := range datas {
		items[item.ID] = item
	}
	return items, nil
}

// 取得技能清單(傳入技能類型)
func GetJsonSkills(skillType SkillType) (map[int]JsonSkill, error) {
	jsonName := JsonName.Skill
	jsonDatas, err := getJsonDic(jsonName)
	if err != nil {
		return nil, err
	}

	skills := make(map[int]JsonSkill)
	for k, v := range jsonDatas {
		jsonSkill, ok := v.(JsonSkill)
		if ok {
			if jsonSkill.Type == skillType {
				skills[k.(int)] = jsonSkill
			}
		} else {
			return nil, fmt.Errorf("%s 取JsonDic時斷言失敗, JsonName: %s", logger.LOG_GameJson, jsonName)
		}

	}
	return skills, nil
}

func GetJsonSkill(id int) (JsonSkill, error) {
	jsonName := JsonName.Skill
	jsonData, err := getJson(JsonName.Skill, id)
	if err != nil {
		log.Errorf("%s 取Json錯誤, JsonName: %s ID: %v", logger.LOG_GameJson, jsonName, id)
		return JsonSkill{}, err
	}
	data, ok := jsonData.(JsonSkill)
	if ok {
		return data, nil
	} else {
		errLog := fmt.Sprintf("%s 取Json時斷言失敗, JsonName: %s ID: %v", logger.LOG_GameJson, jsonName, id)
		log.Errorf(errLog)
		return JsonSkill{}, fmt.Errorf(errLog)
	}
}

func GetJsonSkillsByIDs(ids []int) ([]JsonSkill, error) {
	jsonName := JsonName.Skill
	var skills []JsonSkill

	for _, id := range ids {
		jsonData, err := getJson(jsonName, id)
		if err != nil {
			log.Errorf("%s 取Json錯誤, JsonName: %s ID: %v", logger.LOG_GameJson, jsonName, id)
			return nil, err
		}

		data, ok := jsonData.(JsonSkill)
		if !ok {
			return nil, fmt.Errorf("%s 取Json時斷言失敗, JsonName: %s ID: %v", logger.LOG_GameJson, jsonName, id)
		}

		skills = append(skills, data)
	}

	return skills, nil
}

// 取得隨機技能(傳入技能類型)
func GetRndJsonSkill(skillType SkillType) (JsonSkill, error) {
	normalSkills, err := GetJsonSkills(skillType)
	if err != nil {
		return JsonSkill{}, err
	}

	key := utility.GetRndKeyFromMap(normalSkills)
	data, err := GetJsonSkill(key)
	if err == nil {
		return data, nil
	} else {
		return JsonSkill{}, fmt.Errorf("%s GetJsonSkill錯誤: %v", logger.LOG_GameJson, err)
	}
}
