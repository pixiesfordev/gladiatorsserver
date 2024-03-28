package gameJson

import (
	"encoding/json"
	"fmt"
	"gladiatorsGoModule/utility"
	// "gladiatorsGoModule/logger"
)

type GladiatorJsonData struct {
	ID int `json:"ID"`
}

func (jsonData GladiatorJsonData) UnmarshalJSONData(jsonName string, jsonBytes []byte) (map[int]interface{}, error) {
	var wrapper map[string][]GladiatorJsonData
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

func GetGladiators() ([]GladiatorJsonData, error) {
	datas, err := getJsonDataByName(JsonName.Gladiator)
	if err != nil {
		return nil, err
	}

	var gladiators []GladiatorJsonData
	for _, data := range datas {
		if gladiator, ok := data.(GladiatorJsonData); ok {
			gladiators = append(gladiators, gladiator)
		} else {
			return nil, fmt.Errorf("資料類型不匹配: %T", data)
		}
	}
	return gladiators, nil
}

// 取得隨機鬥士
func GetRndGladiator() (GladiatorJsonData, error) {
	gladiators, err := GetGladiators()
	if err != nil {
		return GladiatorJsonData{}, err
	}
	if len(gladiators) == 0 {
		return GladiatorJsonData{}, fmt.Errorf("鬥士資料為空")
	}
	gladiator, err := utility.GetRandomTFromSlice(gladiators)
	if err != nil {
		return GladiatorJsonData{}, err
	}
	return gladiator, nil
}

func GetGladiatorByID(id int) (GladiatorJsonData, error) {
	gladiators, err := GetGladiators()
	if err != nil {
		return GladiatorJsonData{}, err
	}

	for _, gladiator := range gladiators {
		if gladiator.ID == id {
			return gladiator, nil
		}
	}

	return GladiatorJsonData{}, fmt.Errorf("未找到ID為 %v 的%s資料", id, JsonName.Gladiator)
}
