package gameJson

import (
	"encoding/json"
	"fmt"
	"gladiatorsGoModule/logger"

	log "github.com/sirupsen/logrus"
)

type JsonGameSetting struct {
	ID    string `json:"ID"`
	Value string `json:"Value"`
}

func (jsonData JsonGameSetting) UnmarshalJSONData(jsonName string, jsonBytes []byte) (map[interface{}]interface{}, error) {
	var wrapper map[string][]JsonGameSetting
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

func GetJsonGameSetting(id string) (JsonGameSetting, error) {
	jsonName := JsonName.GameSetting
	jsonData, err := getJson(JsonName.GameSetting, id)
	if err != nil {
		log.Errorf("%s 取Json錯誤, JsonName: %s ID: %v", logger.LOG_GameJson, jsonName, id)
		return JsonGameSetting{}, err
	}
	data, ok := jsonData.(JsonGameSetting)
	if ok {
		return data, nil
	} else {
		return JsonGameSetting{}, fmt.Errorf("%s 取Json時斷言失敗, JsonName: %s ID: %v", logger.LOG_GameJson, jsonName, id)
	}
}
