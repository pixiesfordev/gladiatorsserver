package gameJson

import (
	"encoding/json"
	"fmt"
	// "gladiatorsGoModule/logger"
)

type JsonEquip struct {
	ID int `json:"ID"`
}

func (jsonData JsonEquip) UnmarshalJSONData(jsonName string, jsonBytes []byte) (map[interface{}]interface{}, error) {
	var wrapper map[string][]JsonEquip
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
