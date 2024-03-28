package gameJson

import (
	"encoding/json"
	"fmt"
)

type TraitJsonData struct {
	ID int `json:"ID"`
}

func (jsonData TraitJsonData) UnmarshalJSONData(jsonName string, jsonBytes []byte) (map[int]interface{}, error) {
	var wrapper map[string][]json.RawMessage
	if err := json.Unmarshal(jsonBytes, &wrapper); err != nil {
		return nil, err
	}

	rawDatas, ok := wrapper[jsonName]
	if !ok {
		return nil, fmt.Errorf("找不到key值: %s", jsonName)
	}

	items := make(map[int]interface{})
	for _, rawData := range rawDatas {
		var item TraitJsonData
		if err := json.Unmarshal(rawData, &item); err != nil {
			return nil, err
		}
		items[item.ID] = item
	}
	return items, nil
}

func (traitJson *TraitJsonData) UnmarshalJSON(data []byte) error {
	type Alias TraitJsonData
	aux := &struct {
		RTP     string `json:"RTP"`
		CD      string `json:"CD"`
		Cost    string `json:"Cost"`
		MaxHits string `json:"MaxHits"`
		*Alias
	}{
		Alias: (*Alias)(traitJson), // 使用Alias避免在UnmarshalJSON中呼叫json.Unmarshal時的無限遞迴
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// var err error
	// if aux.RTP != "" {

	// 	rtps, err := utility.Split_FLOAT(aux.RTP, ",")
	// 	if err != nil {
	// 		return err
	// 	}
	// 	spellJson.RTP = rtps
	// }
	// if aux.CD != "" {
	// 	if spellJson.CD, err = strconv.ParseFloat(aux.CD, 64); err != nil {
	// 		return err
	// 	}
	// }
	// if aux.Cost != "" {
	// 	var intVal int64
	// 	if intVal, err = strconv.ParseInt(aux.Cost, 10, 32); err != nil {
	// 		return err
	// 	}
	// 	spellJson.Cost = int32(intVal)
	// }
	// if aux.MaxHits != "" {
	// 	var intVal int64
	// 	if intVal, err = strconv.ParseInt(aux.MaxHits, 10, 32); err != nil {
	// 		return err
	// 	}
	// 	spellJson.MaxHits = int32(intVal)
	// }

	return nil
}

func GetTraits() ([]TraitJsonData, error) {
	datas, err := getJsonDataByName(JsonName.Trait)
	if err != nil {
		return nil, err
	}
	var gladiatorSpells []TraitJsonData
	for _, data := range datas {
		if gladiator, ok := data.(TraitJsonData); ok {
			gladiatorSpells = append(gladiatorSpells, gladiator)
		} else {
			return nil, fmt.Errorf("資料類型不匹配: %T", data)
		}
	}
	return gladiatorSpells, nil
}

func GetTraitByID(id int) (TraitJsonData, error) {
	gladiatorSpells, err := GetTraits()
	if err != nil {
		return TraitJsonData{}, err
	}

	for _, gladiatorSpell := range gladiatorSpells {
		if gladiatorSpell.ID == id {
			return gladiatorSpell, nil
		}
	}
	return TraitJsonData{}, fmt.Errorf("未找到ID為 %v 的%s資料", id, JsonName.Trait)
}
