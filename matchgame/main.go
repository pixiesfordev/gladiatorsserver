package main

import (
	"gladiatorsGoModule/k8s"
	"gladiatorsGoModule/setting"
	logger "matchgame/logger"
	"strconv"

	// gSetting "matchgame/setting"

	log "github.com/sirupsen/logrus"

	"flag"

	agonesv1 "agones.dev/agones/pkg/apis/agones/v1"
	serverSDK "agones.dev/agones/pkg/sdk"
	"agones.dev/agones/pkg/util/signals"

	// "fmt"
	"gladiatorsGoModule/gameJson"
	mongo "gladiatorsGoModule/mongo"
	"matchgame/game"
	"os"
	"time"
)

// 環境版本
const (
	ENV_DEV     = "Dev"
	ENV_RELEASE = "Release"
)

var Env string // 環境版本

var PodName string // Pod名稱

// 初始化日誌設置
func initLogger() {
	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

// 從環境變量獲取配置
func getEnvConfig() (*string, error) {
	port := flag.String("port", "7654", "The port to listen to tcp traffic on")
	if ep := os.Getenv("PORT"); ep != "" {
		port = &ep
	}

	// 檢查必要的環境變量
	if imgVer := os.Getenv("ImgVer"); imgVer != "" {
		log.Infof("%s Image版本為: %s", logger.LOG_Main, imgVer)
	} else {
		log.Errorf("%s 取不到環境變數: ImgVer", logger.LOG_Main)
	}

	if game.Mode = os.Getenv("Mode"); game.Mode != "" {
		log.Infof("%s Mode為: %s", logger.LOG_Main, game.Mode)
	} else {
		game.Mode = "standard"
		log.Errorf("%s 取不到環境變數: Mode", logger.LOG_Main)
	}

	if PodName = os.Getenv("PodName"); PodName != "" {
		log.Infof("%s PodName為: %s", logger.LOG_Main, PodName)
	} else {
		log.Errorf("%s 取不到環境變數: PodName", logger.LOG_Main)
	}

	Env = *flag.String("Version", "Dev", "version setting")
	if ep := os.Getenv("Version"); ep != "" {
		Env = ep
	}

	return port, nil
}

// 獲取MongoDB認證信息
func getMongoCredentials() (string, string, string, string) {
	return os.Getenv("MongoAPIPublicKey"),
		os.Getenv("MongoAPIPrivateKey"),
		os.Getenv("MongoUser"),
		os.Getenv("MongoPW")
}

// 輸出房間初始化日誌
func logRoomInitInfo(packID int64, podName, nodeName string, playerIDs [setting.PLAYER_NUMBER]string, dbMapID, roomName string, port interface{}) {
	log.Infof("%s ==============開始初始化房間==============", logger.LOG_Main)
	log.Infof("%s packID: %v", logger.LOG_Main, packID)
	log.Infof("%s podName: %v", logger.LOG_Main, podName)
	log.Infof("%s nodeName: %v", logger.LOG_Main, nodeName)
	log.Infof("%s PlayerIDs: %s", logger.LOG_Main, playerIDs)
	log.Infof("%s dbMapID: %s", logger.LOG_Main, dbMapID)
	log.Infof("%s roomName: %s", logger.LOG_Main, roomName)
	log.Infof("%s Port: %v", logger.LOG_Main, port)
	log.Infof("%s Get Info Finished", logger.LOG_Main)
}

func main() {
	initLogger()
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Main Crash: %v", r)
		}
	}()

	log.Infof("%s ==============MATCHGAME 啟動==============", logger.LOG_Main)

	port, err := getEnvConfig()
	if err != nil {
		log.Errorf("%s 獲取環境配置失敗: %v", logger.LOG_Main, err)
		return
	}

	if game.Mode != "non-agones" { // non-agones模式下不要呼叫初始化Agones, 也不要偵聽agones訊號
		err := game.InitAgones()
		if err != nil {
			log.Errorf("%s %s", logger.LOG_Main, err)
		}
		go signalListen()
	}

	// 共用的初始化邏輯
	mongoAPIPublicKey, mongoAPIPrivateKey, mongoUser, mongoPW := getMongoCredentials()
	initMonogo(mongoAPIPublicKey, mongoAPIPrivateKey, mongoUser, mongoPW)
	InitGameJson() // 初始化遊戲Json資料

	roomCreatedChan := make(chan struct{})
	var packID = int64(0)

	if game.Mode == "standard" {
		var lobbyPodName string
		var err error
		roomInit := false

		game.AgonesSDK.WatchGameServer(func(gs *serverSDK.GameServer) {

			// log.Infof("%s 遊戲房狀態 %s", logger.LOG_Main, gs.Status.State)
			defer func() {
				if err := recover(); err != nil {
					log.Errorf("%s 遊戲崩潰: %v.\n", logger.LOG_Main, err)
					game.AgonesSDK.Shutdown()
				}
			}()

			if !roomInit && gs.ObjectMeta.Labels["RoomName"] != "" {
				log.Infof("%s 開始新局遊戲", logger.LOG_Main)

				lobbyPodName = gs.ObjectMeta.Labels["LobbyPodName"]

				playerIDs := [setting.PLAYER_NUMBER]string{}

				// 從Agones取得房間資訊
				dbMapID := gs.ObjectMeta.Labels["DBMapID"]
				roomInit = true
				packID, err = strconv.ParseInt(gs.ObjectMeta.Labels["PackID"], 10, 64)
				if err != nil {
					log.Errorf("%s strconv.ParseInt packID錯誤: %v", logger.LOG_Main, err)
				}
				roomName := gs.ObjectMeta.Labels["RoomName"]
				podName := gs.ObjectMeta.Name
				nodeName := os.Getenv("NodeName")
				logRoomInitInfo(packID, podName, nodeName, playerIDs, dbMapID, roomName, int(gs.Status.Ports[0].Port))

				game.InitGameRoom(dbMapID, playerIDs, roomName, gs.Status.Address, int(gs.Status.Ports[0].Port), podName, nodeName, lobbyPodName, roomCreatedChan)
				game.SetServerState(agonesv1.GameServerStateAllocated) // 設定房間為Allocated(agones應該會在WatchGameServer後自動設定為Allocated但這邊還是主動設定)
				log.Infof("%s GameServer狀態為: %s", logger.LOG_Main, gs.Status.State)
				log.Infof("%s ==============初始化房間完成==============", logger.LOG_Main)
			} else {
				if lobbyPodName != "" && gs.ObjectMeta.Labels["LobbyPodName"] != "" && lobbyPodName != gs.ObjectMeta.Labels["LobbyPodName"] {
					log.Errorf("%s Agones has allocate error in parelle", logger.LOG_Main)
				}
			}
		})
	} else if game.Mode == "non-agones" { // non-agones模式

		log.Infof("%s 開始新局遊戲", logger.LOG_Main)

		go func() {

			myPort, parsePortErr := strconv.ParseInt(*port, 10, 32)
			if parsePortErr != nil {
				log.Errorf("%s parse Port錯誤: %v", logger.LOG_Main, parsePortErr)
			}

			// 處理外部IP和端口
			tcpIP := getAndWaitForExternalIP()
			setExternalIPandPort(tcpIP, int(myPort))

			// 初始化房間資訊
			lobbyPodName := ""
			playerIDs := [setting.PLAYER_NUMBER]string{}

			// 從DB取得房間資訊
			dbGameState, err := mongo.GetDocByID[mongo.DBGameState](mongo.Col.GameSetting, "GameState")
			if err != nil {
				log.Errorf("%s InitGameRoom時取DBGameState資料發生錯誤: %v", logger.LOG_Main, err)
			}

			dbMapID := dbGameState.MatchgameTestverMapID
			roomName := dbGameState.MatchgameTestverRoomName
			nodeName := os.Getenv("NodeName")

			logRoomInitInfo(packID, PodName, nodeName, playerIDs, dbMapID, roomName, port)
			game.InitGameRoom(dbMapID, playerIDs, roomName, "", int(myPort), PodName, nodeName, lobbyPodName, roomCreatedChan)

			log.Infof("%s ==============初始化房間完成==============", logger.LOG_Main)
		}()
	}

	stopChan := make(chan struct{})
	endGameChan := make(chan struct{})
	if game.Mode != "non-agones" {
		game.SetServerState(agonesv1.GameServerStateReady) // 設定房間為Ready(才會被Matchmaker分配玩家進來)
		go game.AgonesHealthPin(stopChan)                  // Agones伺服器健康檢查
	}
	game.InitGame() // 初始化遊戲
	<-roomCreatedChan
	close(roomCreatedChan)
	log.Infof("%s ==============Room資料設定完成==============", logger.LOG_Main)

	// 開啟連線
	src := ":" + *port
	go openConnectTCP(stopChan, src)
	// go openConnectUDP(stopChan, src)
	go game.RunGameTimer(stopChan) // 開始遊戲房計時器

	select {
	case <-stopChan:
		// 錯誤發生寫入Log
		// FirebaseFunction.DeleteGameRoom(RoomName)
		log.Infof("%s game stop chan", logger.LOG_Main)
		if game.Mode != "non-agones" { // non-agones模式下不使用agones服務
			game.ShutdownAgonesServer()
		}

		return
	case <-endGameChan:
		// 遊戲房關閉寫入Log
		// FirebaseFunction.DeleteGameRoom(RoomName)
		log.Infof("%s End game chan", logger.LOG_Main)
		if game.Mode != "non-agones" { // non-agones模式下不使用agones服務
			game.DelayShutdownServer(60*time.Second, stopChan)
		}
	}
	<-stopChan

	if game.Mode != "non-agones" { // non-agones模式下不使用agones服務
		game.ShutdownAgonesServer() // 關閉Server
	}

}

// 取Loadbalancer分配給此pod的對外
func getExternalIP(_serviceName string) (string, error) {
	ip, err := k8s.GetLoadBalancerExternalIP(setting.NAMESPACE_GAMESERVER, _serviceName)
	if err != nil {
		log.Errorf("%s GetLoadBalancerExternalIP error: %v.\n", logger.LOG_Main, err)
	}
	return ip, err
}

// 寫入對外ID到DB中
func setExternalIPandPort(tcpIP string, port int) {
	log.Infof("%s 開始寫入對外IP到DB tcpIP:%s port:%v.\n", logger.LOG_Main, tcpIP, port)
	// 設定要更新的資料
	updateData := struct {
		MatchgameTestverTcpIp string `bson:"matchgameTestverTcpIp"`
		MatchgameTestverPort  int    `bson:"matchgameTestverPort"`
	}{
		MatchgameTestverTcpIp: tcpIP,
		MatchgameTestverPort:  port,
	}
	log.Infof("更新數據內容: %+v", updateData)
	// 更新資料
	result, err := mongo.UpdateDocByStruct(mongo.Col.GameSetting, "GameState", updateData)
	if err != nil {
		log.Errorf("%s SetExternalID失敗: %v", logger.LOG_Main, err)
		return
	}
	log.Infof("%s 寫入對外IP完成 %+v \n", logger.LOG_Main, result)
}

// 初始化遊戲Json資料
func InitGameJson() {
	log.Infof("%s 開始初始化GameJson", logger.LOG_Main)
	err := gameJson.Init(Env)
	if err != nil {
		log.Errorf("%s 初始化GameJson失敗: %v", logger.LOG_Main, err)
		return
	}
}

// 初始化MongoDB設定
func initMonogo(mongoAPIPublicKey string, mongoAPIPrivateKey string, user string, pw string) {
	log.Infof("%s 初始化mongo開始", logger.LOG_Main)
	mongo.Init(mongo.InitData{
		Env:           Env,
		APIPublicKey:  mongoAPIPublicKey,
		APIPrivateKey: mongoAPIPrivateKey,
	}, user, pw)
	log.Infof("%s 初始化mongo完成", logger.LOG_Main)
}

// 偵測SIGTERM/SIGKILL的終止訊號，偵測到就刪除遊戲房資料並寫log
func signalListen() {
	ctx, _ := signals.NewSigKillContext()
	<-ctx.Done()
	// FirebaseFunction.DeleteGameRoom(documentID)
	log.Infof("%s Exit signal received. Shutting down.", logger.LOG_Main)
	os.Exit(0)
}

// 新增的輔助函數：等待並獲取外部IP
func getAndWaitForExternalIP() string {
	log.Infof("%s 取Loadbalancer分配給此pod的對外IP.\n", logger.LOG_Main)
	var tcpIP string
	for tcpIP == "" {
		time.Sleep(1 * time.Second)
		getTcpIP, err := getExternalIP(setting.MATCHGAME_TESTVER_TCP)
		if err != nil {
			break
		}
		if getTcpIP != "" {
			tcpIP = getTcpIP
			log.Infof("%s 取得對外TCP IP成功: %s .\n", logger.LOG_Main, tcpIP)
			break
		}
	}
	return tcpIP
}
