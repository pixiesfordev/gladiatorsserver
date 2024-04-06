package gameJson

import (
	"encoding/json"
	"fmt"
	"gladiatorsGoModule/logger"
	"gladiatorsGoModule/utility"

	log "github.com/sirupsen/logrus"
)

type JsonBribe struct {
	ID   int `json:"ID"`
	Cost int `json:"Cost"`
}

func (jsonData JsonBribe) UnmarshalJSONData(jsonName string, jsonBytes []byte) (map[interface{}]interface{}, error) {
	var wrapper map[string][]JsonBribe
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

func GetJsonBribes() (map[int]JsonBribe, error) {
	jsonName := JsonName.Bribe
	jsonDatas, err := getJsonDic(jsonName)
	if err != nil {
		return nil, err
	}
	datas := make(map[int]JsonBribe)
	for _, v := range jsonDatas {
		json, ok := v.(JsonBribe)
		if ok {
			datas[json.ID] = json
		} else {
			return nil, fmt.Errorf("%s 取JsonDic時斷言失敗, JsonName: %s", logger.LOG_GameJson, jsonName)
		}
	}
	return datas, nil
}

func GetJsonBribe(id int) (JsonBribe, error) {
	jsonName := JsonName.Bribe
	jsonData, err := getJson(JsonName.Bribe, id)
	if err != nil {
		log.Errorf("%s 取Json錯誤, JsonName: %s ID: %v", logger.LOG_GameJson, jsonName, id)
		return JsonBribe{}, err
	}
	data, ok := jsonData.(JsonBribe)
	if ok {
		return data, nil
	} else {
		return JsonBribe{}, fmt.Errorf("%s 取Json時斷言失敗, JsonName: %s ID: %v", logger.LOG_GameJson, jsonName, id)
	}
}

func GetRndJsonBribe() (JsonBribe, error) {
	jsonName := JsonName.Bribe
	jsonDatas, err := getJsonDic(jsonName)
	if err != nil {
		return JsonBribe{}, err
	}
	key := utility.GetRndKeyFromMap(jsonDatas)
	data, err := GetJsonBribe(key.(int))
	if err == nil {
		return data, nil
	} else {
		return JsonBribe{}, fmt.Errorf("%s GetJsonBribe錯誤: %v", logger.LOG_GameJson, err)
	}
}
