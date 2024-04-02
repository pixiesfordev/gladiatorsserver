package gameJson

import (
	"context"
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
	bucketName := "gladiators_gamejson_dev"
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
				SetJsonDic(jsonName, data, JsonGameSetting{})
			case JsonName.Gladiator:
				SetJsonDic(jsonName, data, JsonGladiator{})
			case JsonName.Trait:
				SetJsonDic(jsonName, data, TraitJsonData{})
			case JsonName.Skill:
				SetJsonDic(jsonName, data, JsonSkill{})
			case JsonName.Equip:
				SetJsonDic(jsonName, data, JsonEquip{})
			default:
				log.Errorf("%s 未定義的jsonName: %v", logger.LOG_GameJson, jsonName)
			}

		}
	}
	return nil
}

// jsonDic的結構為jsonDic[jsonName][ID]
var jsonDic = make(map[string]map[interface{}]interface{})

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

type JsonUnmarshaler interface {
	UnmarshalJSONData(jsonName string, jsonData []byte) (map[interface{}]interface{}, error)
}

func SetJsonDic(jsonName string, jsonData []byte, unmarshaler JsonUnmarshaler) error {
	items, err := unmarshaler.UnmarshalJSONData(jsonName, jsonData)
	if err != nil {
		log.Printf("%s表Unmarshal失敗: %v", jsonName, err)
		return err
	}
	jsonDic[jsonName] = items
	log.Printf("設定Json表(%s) %v筆資料", jsonName, len(items))
	return nil
}

// 傳入Json名稱取得對應JsonMap資料
func getJsonDic(jsonName string) (map[interface{}]interface{}, error) {
	dic, exists := jsonDic[jsonName]
	if !exists {
		return nil, fmt.Errorf("jsonDic中未找到 %s 的資料", jsonName)
	}
	return dic, nil
}

// 傳入Json名稱取得對應JsonMap資料
func getJson(jsonName string, id interface{}) (interface{}, error) {
	dic, exists := jsonDic[jsonName]
	if !exists {
		return nil, fmt.Errorf("jsonDic中未找到資料 JsonName: %s", jsonName)
	}
	data, ok := dic[id]
	if !ok {
		return nil, fmt.Errorf("jsonDic中未找到資料 JsonName: %s ID: %v", jsonName, id)
	}
	return data, nil
}
