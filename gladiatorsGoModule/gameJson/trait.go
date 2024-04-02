package gameJson

import (
	"encoding/json"
	"fmt"
)

type TraitJsonData struct {
	ID int `json:"ID"`
}

func (jsonData TraitJsonData) UnmarshalJSONData(jsonName string, jsonBytes []byte) (map[interface{}]interface{}, error) {
	var wrapper map[string][]json.RawMessage
	if err := json.Unmarshal(jsonBytes, &wrapper); err != nil {
		return nil, err
	}

	rawDatas, ok := wrapper[jsonName]
	if !ok {
		return nil, fmt.Errorf("找不到key值: %s", jsonName)
	}

	items := make(map[interface{}]interface{})
	for _, rawData := range rawDatas {
		var item TraitJsonData
		if err := json.Unmarshal(rawData, &item); err != nil {
			return nil, err
		}
		items[item.ID] = item
	}
	return items, nil
}
