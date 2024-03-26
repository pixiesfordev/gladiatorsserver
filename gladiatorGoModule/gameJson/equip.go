package gameJson

import (
	"encoding/json"
	"fmt"
	// "gladiatorsGoModule/logger"
)

// GladiatorEXP JSON
type EquipJsonData struct {
	ID int `json:"ID"`
}

func (jsonData EquipJsonData) UnmarshalJSONData(jsonName string, jsonBytes []byte) (map[int]interface{}, error) {
	var wrapper map[string][]EquipJsonData
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

func GetEquips() ([]EquipJsonData, error) {
	datas, err := getJsonDataByName(JsonName.Equip)
	if err != nil {
		return nil, err
	}

	var gladiatorEXPs []EquipJsonData
	for _, data := range datas {
		if gladiatorEXP, ok := data.(EquipJsonData); ok {
			gladiatorEXPs = append(gladiatorEXPs, gladiatorEXP)
		} else {
			return nil, fmt.Errorf("資料類型不匹配: %T", data)
		}
	}
	return gladiatorEXPs, nil
}
