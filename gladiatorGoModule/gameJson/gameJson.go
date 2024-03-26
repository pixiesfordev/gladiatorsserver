package gameJson

import (
	"context"
	"errors"
	"gladiatorsGoModule/setting"
	"io"
	"strings"

	"fmt"
	"gladiatorsGoModule/logger"

	"cloud.google.com/go/storage"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
)

// 初始化JsonMap
func Init(env string) error {
	gcpProjectID, ok := setting.EnvGCPProject[env]
	if !ok {
		// log.Errorf("%s env錯誤: %s", logger.LOG_GameJson, env)
		return fmt.Errorf("%s evn名稱錯誤: %v", logger.LOG_GameJson, env)
	}
	log.Infof("%s gcpProjectID: %s", logger.LOG_GameJson, gcpProjectID)

	ctx := context.Background()

	// 初始化 GCS 客戶端
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("%s GCS初始化錯誤: %v", logger.LOG_GameJson, err)
	}

	// 設定bucket和object前綴
	bucketName := "gladiators_gamejson_dev3"
	prefix := "" // 如果所有的json都在根目錄，就用空字串就可以

	bucket := client.Bucket(bucketName)
	// 創建一個列舉object的查詢
	query := &storage.Query{Prefix: prefix}

	// 執行查詢
	item := bucket.Objects(ctx, query)
	index := 0
	for {
		attrs, err := item.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("%s ListBucket: %v", logger.LOG_GameJson, err)
		}

		// 檢查是否為.json檔案
		if len(attrs.Name) > 5 && attrs.Name[len(attrs.Name)-5:] == ".json" {
			index++
			object := bucket.Object(attrs.Name)
			reader, err := object.NewReader(ctx)
			if err != nil {
				return fmt.Errorf("%s Failed to read object: %v", logger.LOG_GameJson, err)
			}

			data, err := io.ReadAll(reader)
			reader.Close()
			if err != nil {
				return fmt.Errorf("%s Failed to read data: %v", logger.LOG_GameJson, err)
			}
			jsonName := strings.TrimSuffix(attrs.Name, ".json")
			// fmt.Printf("%s File: %s Data: %s \n", logger.LOG_GameJson, attrs.Name, data)
			switch jsonName {
			case JsonName.GameSetting:
				SetJsonDic(jsonName, data, GameSettingJsonData{})
			case JsonName.Gladiator:
				SetJsonDic(jsonName, data, GladiatorJsonData{})
			case JsonName.Trait:
				SetJsonDic(jsonName, data, TraitJsonData{})
			case JsonName.Skill:
				SetJsonDic(jsonName, data, SkillJsonData{})
			case JsonName.Equip:
				SetJsonDic(jsonName, data, EquipJsonData{})
			default:
				log.Errorf("%s 未定義的jsonName: %v", logger.LOG_GameJson, jsonName)
				return errors.New("未定義的jsonName")
			}

		}
	}
	return nil
}

// jsonDic的結構為jsonDic[jsonName][ID]
var jsonDic = make(map[string]map[string]interface{})

type JsonNameStruct struct {
	GameSetting string
	Gladiator   string
	Equip       string
	Skill       string
	Trait       string
}

// Json名稱列表
var JsonName = JsonNameStruct{
	GameSetting: "GameSetting",
	Gladiator:   "Gladiator",
	Equip:       "Equip",
	Skill:       "Skill",
	Trait:       "Trait",
}

// 傳入Json名稱取得對應JsonMap資料
func getJsonDataByName(name string) (map[string]interface{}, error) {
	data, exists := jsonDic[name]
	if !exists {
		return nil, fmt.Errorf("jsonDic中未找到 %s 的資料", name)
	}
	return data, nil
}

type JsonUnmarshaler[K comparable] interface {
	UnmarshalJSONData(jsonName string, jsonData []byte) (map[K]interface{}, error)
}

func SetJsonDic[K comparable](jsonName string, jsonData []byte, unmarshaler JsonUnmarshaler[K]) error {
	items, err := unmarshaler.UnmarshalJSONData(jsonName, jsonData)
	if err != nil {
		log.Printf("%s表Unmarshal失敗: %v", jsonName, err)
		return err
	}
	// 這裡使用interface{}以便於處理不同類型的map，你可能需要根據實際情況進行調整
	jsonDic := make(map[string]interface{})
	jsonDic[jsonName] = items
	log.Printf("設定Json資料: %s", jsonName)
	return nil
}

// // 傳入Json將並轉為對應struct資料並存入jsonDic中, jsonDic的結構為jsonDic[jsonName][ID]
// func SetJsonDic(jsonName string, jsonData []byte) error {

// 	var unmarshaler JsonUnmarshaler
// 	switch jsonName {
// 	case JsonName.GameSetting:
// 		unmarshaler = GameSettingJsonData{}
// 	case JsonName.Gladiator:
// 		unmarshaler = GladiatorJsonData{}
// 	case JsonName.Trait:
// 		unmarshaler = TraitJsonData{}
// 	case JsonName.Skill:
// 		unmarshaler = SkillJsonData{}
// 	case JsonName.Equip:
// 		unmarshaler = EquipJsonData{}
// 	default:
// 		log.Errorf("%s 未定義的jsonName: %v", logger.LOG_GameJson, jsonName)
// 		return errors.New("未定義的jsonName")
// 	}
// 	items, err := unmarshaler.UnmarshalJSONData(jsonName, jsonData)
// 	if err != nil {
// 		log.Errorf("%s %s表Unmarshal失敗: %v", logger.LOG_GameJson, jsonName, err)
// 		return err
// 	}
// 	jsonDic[jsonName] = items
// 	log.Infof("%s 設定Json資料: %s", logger.LOG_GameJson, jsonName)
// 	return nil
// }
