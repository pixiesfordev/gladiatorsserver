package game

import (
	"gladiatorsGoModule/env"
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/logger"
	"gladiatorsGoModule/mongo"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

var Env string                      // 環境版本
var SelfPodName string              // K8s上所屬的Pod名稱
var MyUsher *Usher                  // 配房者
var MyDBMaps map[string]mongo.DBMap // 地圖資料

// InitGame 初始化遊戲
func InitGame() {
	log.Infof("%v 初始化遊戲", logger.LOG_Main)
	// 初始化Usher
	MyUsher = NewUsher()
	Env = env.GetEnv("Env", "", "", false)
	SelfPodName = env.GetEnv("PodName", "", "", false)
	gameJson.Init(Env) // 初始化 GameJson

	// 初始化地圖資料
	MyDBMaps = make(map[string]mongo.DBMap)
	if err := initDBMaps(); err != nil {
		log.Errorf("%s InitGame initDBMaps error: %v", logger.LOG_Main, err)
	}

	InitAgones() // 初始化Agones
	log.Infof("%v 初始化遊戲完成", logger.LOG_Main)
}

// initDBMaps 初始化地圖資料
func initDBMaps() error {
	var dbMaps []mongo.DBMap
	filter := bson.M{"enable": true}
	err := mongo.GetDocsByFilter(mongo.Col.Map, filter, &dbMaps)
	if err != nil {
		log.Errorf("%s 取得地圖資料失敗: %v", logger.LOG_Main, err)
		return err
	}

	for _, dbMap := range dbMaps {
		MyDBMaps[dbMap.ID] = dbMap
	}

	log.Infof("%s 取得 %d 個啟用中的地圖資料", logger.LOG_Main, len(MyDBMaps))
	return nil
}

// GetDBMap 取得地圖資料
func GetDBMap(mapID string) (mongo.DBMap, bool) {
	dbMap, ok := MyDBMaps[mapID]
	return dbMap, ok
}
