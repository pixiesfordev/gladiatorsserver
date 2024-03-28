package gameJson

import (
	"encoding/json"
	"fmt"
)

type SkillJsonData struct {
	ID int `json:"ID"`
}

func (jsonData SkillJsonData) UnmarshalJSONData(jsonName string, jsonBytes []byte) (map[int]interface{}, error) {
	var wrapper map[string][]SkillJsonData
	if err := json.Unmarshal(jsonBytes, &wrapper); err != nil {
		return nil, err
	}

	datas, ok := wrapper[jsonName]
	if !ok {
		return nil, fmt.Errorf("找不到key值: %s", jsonName)
	}

	items := make(map[int]interface{})
	for _, item := range datas {
		items[item.ID] = item
	}
	return items, nil
}

func GetSkills() ([]SkillJsonData, error) {
	datas, err := getJsonDataByName(JsonName.Skill) // Assuming you have JsonName.GladiatorSkin defined
	if err != nil {
		return nil, err
	}

	var gladiatorSkins []SkillJsonData
	for _, data := range datas {
		if gladiatorSkin, ok := data.(SkillJsonData); ok {
			gladiatorSkins = append(gladiatorSkins, gladiatorSkin)
		} else {
			return nil, fmt.Errorf("資料類型不匹配: %T", data)
		}
	}
	return gladiatorSkins, nil
}

func GetSkillByID(id int) (SkillJsonData, error) {
	gladiatorSkins, err := GetSkills()
	if err != nil {
		return SkillJsonData{}, err
	}

	for _, gladiatorSkin := range gladiatorSkins {
		if gladiatorSkin.ID == id {
			return gladiatorSkin, nil
		}
	}

	return SkillJsonData{}, fmt.Errorf("未找到ID為 %v 的%s資料", id, JsonName.Skill)
}
