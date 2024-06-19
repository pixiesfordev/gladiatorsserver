package gameJson

import (
	"encoding/json"
	"fmt"
	"gladiatorsGoModule/logger"
	"gladiatorsGoModule/utility"

	log "github.com/sirupsen/logrus"
)

type JsonSkill struct {
	ID         int    `json:"ID"`
	Activation string `json:"Activation"`
	Initiative int    `json:"Initiative"`
	Vigor      int    `json:"Vigor"`
	Divine     string `json:"Divine"`
}

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

func GetJsonSkills() (map[int]JsonSkill, error) {
	jsonName := JsonName.Skill
	jsonDatas, err := getJsonDic(jsonName)
	if err != nil {
		return nil, err
	}
	datas := make(map[int]JsonSkill)
	for _, v := range jsonDatas {
		json, ok := v.(JsonSkill)
		if ok {
			datas[json.ID] = json
		} else {
			return nil, fmt.Errorf("%s 取JsonDic時斷言失敗, JsonName: %s", logger.LOG_GameJson, jsonName)
		}
	}
	return datas, nil
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
		return JsonSkill{}, fmt.Errorf("%s 取Json時斷言失敗, JsonName: %s ID: %v", logger.LOG_GameJson, jsonName, id)
	}
}

func GetRndJsonSkill() (JsonSkill, error) {
	jsonName := JsonName.Skill
	jsonDatas, err := getJsonDic(jsonName)
	if err != nil {
		return JsonSkill{}, err
	}
	key := utility.GetRndKeyFromMap(jsonDatas)
	data, err := GetJsonSkill(key.(int))
	if err == nil {
		return data, nil
	} else {
		return JsonSkill{}, fmt.Errorf("%s GetJsonSkill錯誤: %v", logger.LOG_GameJson, err)
	}
}
