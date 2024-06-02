package gameJson

import (
	"encoding/json"
	"fmt"
	"gladiatorsGoModule/logger"
	"gladiatorsGoModule/utility"

	log "github.com/sirupsen/logrus"
)

type JsonGladiator struct {
	ID         int     `json:"ID"`
	HP         int     `json:"HP"`
	STR        int     `json:"STR"`
	DEF        int     `json:"DEF"`
	MDEF       int     `json:"MDEF"`
	CRIT       float64 `json:"CRIT"`
	VigorRegen float64 `json:"VigorRegen"`
	Knockback  int     `json:"Knockback"`
	INIT       int     `json:"INIT"`
	Speed      int     `json:"Speed"`
}

func (jsonData JsonGladiator) UnmarshalJSONData(jsonName string, jsonBytes []byte) (map[interface{}]interface{}, error) {
	var wrapper map[string][]JsonGladiator
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

func GetJsonGladiators() (map[int]JsonGladiator, error) {
	jsonName := JsonName.Gladiator
	jsonDatas, err := getJsonDic(jsonName)
	if err != nil {
		return nil, err
	}
	datas := make(map[int]JsonGladiator)
	for _, v := range jsonDatas {
		json, ok := v.(JsonGladiator)
		if ok {
			datas[json.ID] = json
		} else {
			return nil, fmt.Errorf("%s 取JsonDic時斷言失敗, JsonName: %s", logger.LOG_GameJson, jsonName)
		}
	}
	return datas, nil
}

func GetJsonGladiator(id int) (JsonGladiator, error) {
	jsonName := JsonName.Gladiator
	jsonData, err := getJson(JsonName.Gladiator, id)
	if err != nil {
		log.Errorf("%s 取Json錯誤, JsonName: %s ID: %v", logger.LOG_GameJson, jsonName, id)
		return JsonGladiator{}, err
	}
	data, ok := jsonData.(JsonGladiator)
	if ok {
		return data, nil
	} else {
		return JsonGladiator{}, fmt.Errorf("%s 取Json時斷言失敗, JsonName: %s ID: %v", logger.LOG_GameJson, jsonName, id)
	}
}

func GetRndJsonGladiator() (JsonGladiator, error) {
	jsonName := JsonName.Gladiator
	jsonDatas, err := getJsonDic(jsonName)
	if err != nil {
		return JsonGladiator{}, err
	}
	key := utility.GetRndKeyFromMap(jsonDatas)
	data, err := GetJsonGladiator(key.(int))
	if err == nil {
		return data, nil
	} else {
		return JsonGladiator{}, fmt.Errorf("%s GetJsonGladiator錯誤: %v", logger.LOG_GameJson, err)
	}
}
