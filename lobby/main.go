package main

import (
	"fmt"
	"gladiatorsGoModule/env"
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/mongo"
	"lobby/config"
	"lobby/logger"
	"lobby/middleware"
	"lobby/restfulAPI"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func main() {
	logger.InitLogger()
	log.Infof("%s ============= Lobby 啟動 ==============", logger.LOG_Main)

	// 設定環境版本
	config.Set(env.GetEnv("Env", "Dev", "", false))

	// 初始化 MongoDB 設定
	initMongo(
		env.GetEnv("MongoAPIPublicKey", "", "", false),
		env.GetEnv("MongoAPIPrivateKey", "", "", false),
		env.GetEnv("MongoUser", "", "", false),
		env.GetEnv("MongoPW", "", "", false),
	)
	// 初始化 GameJson
	gameJson.Init(config.Env())

	// 建立路由
	initRouter(env.GetEnv("PORT_HTTPS", "", "", false))

	// 初始化Agones
	initAgones()

	// 設定Loadbalance分配的對外IP
	tcpPort := env.GetEnv("PORT_TCP", "", "", false)
	err := podSet(tcpPort)
	if err != nil {
		log.Errorf("%v podSet失敗: %v", logger.LOG_Main, err)
		select {} // 停止主程式避免程式結束後又被K8s自動重啟
	}

	// 建立[TCP]socket連線
	initTcp(tcpPort)

	log.Infof("%s ============= Lobby 啟動完成 ==============", logger.LOG_Main)

}

// 初始化 MongoDB 設定
func initMongo(mongoAPIPublicKey string, mongoAPIPrivateKey string, user string, pw string) {
	log.Infof("%s 初始化 mongo 開始", logger.LOG_Main)
	mongo.Init(mongo.InitData{
		Env:           config.Env(),
		APIPublicKey:  mongoAPIPublicKey,
		APIPrivateKey: mongoAPIPrivateKey,
	}, user, pw)
	log.Infof("%s 初始化 mongo 完成", logger.LOG_Main)
}

// initRouter 建立路由
func initRouter(port string) {
	router := mux.NewRouter()
	// GET
	router.HandleFunc("/game/servertime", restfulAPI.ServerTime).Methods("GET")
	router.HandleFunc("/game/gamestate", restfulAPI.GameState).Methods("GET")
	// POST
	router.HandleFunc("/game/signup", restfulAPI.Signup).Methods("POST")
	router.HandleFunc("/game/signin", restfulAPI.Signin).Methods("POST")

	// 使用 Middlewares
	router.Use(
		middleware.RecoveryMiddleware,
		middleware.RequestInfoMiddleware,
	)
	port = fmt.Sprintf(":%s", port)
	// 啟動路由伺服器於 Goroutine 中
	go func() {
		log.Infof("%s 路由初始化完成", logger.LOG_Main)
		if err := http.ListenAndServe(port, router); err != nil {
			log.Fatalf("%s 路由啟動失敗: %v", logger.LOG_Main, err)
		}
	}()
}

func podSet(port string) error {
	// 取Loadbalancer分配給此pod的對外IP並寫入資料庫
	log.Infof("%s 取Loadbalancer分配給此pod的對外IP.\n", logger.LOG_Main)
	for {
		// 因為pod啟動後Loadbalancer並不會立刻就分配好ip(會有延遲) 所以每5秒取一次 直到取到ip才往下跑
		time.Sleep(1 * time.Second) // 每5秒取一次ip
		ip, getIPErr := getExternalIP()
		if getIPErr != nil {
			return fmt.Errorf("%v getExternalIP失敗: %v", logger.LOG_Main, getIPErr)
		}
		if ip != "" {
			log.Infof("%s 取得對外IP成功: %s", logger.LOG_Main, ip)
			err := updatePodDataToDB(ip, port) // 寫入對外ID到DB中
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}

// updatePodDataToDB 更新Pod資料到DB中
func updatePodDataToDB(ip, port string) error {
	log.Infof("%s 開始更新Pod資料到DB中", logger.LOG_Main)
	intPort, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("%v strconv.Atoi(port)錯誤: %v", logger.LOG_Main, err)
	}
	updateData := struct {
		LobbyIP   string `bson:"lobbyIP"`
		LobbyPort int    `bson:"lobbyPort"`
	}{
		LobbyIP:   ip,
		LobbyPort: intPort,
	}
	_, err = mongo.UpdateDocByStruct(mongo.Col.GameSetting, "GameState", updateData)
	if err != nil {
		return fmt.Errorf("%s mongo.UpdateDocByStruct失敗: %v", logger.LOG_Main, err)
	}
	log.Infof("%s 更新Pod資料到DB成功", logger.LOG_Main)
	return nil
}
