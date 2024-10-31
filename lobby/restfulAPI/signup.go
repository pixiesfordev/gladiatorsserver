package restfulAPI

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gladiatorsGoModule/mongo"
	"gladiatorsGoModule/setting"
	"lobby/logger"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DB 寫入資料
type (
	signupData struct {
		AuthType   string `json:"authType"`
		AuthData   string `json:"authData"`
		DeviceType string `json:"deviceType"`
		DeviceUID  string `json:"deviceUID"`
	}
)

// [POSTER] /game/signup
func Signup(w http.ResponseWriter, r *http.Request) {

	var data signupData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 建立 Log Instance
	logData := logger.NewSignupData()

	// 產生 PlayerID
	playerID := primitive.NewObjectID().Hex()

	// 產生Auth資料
	var authDatas map[string]string
	// Guest:裝置UID
	switch data.AuthType {
	case setting.AUTH_GUEST:
		authDatas = map[string]string{data.AuthType: data.AuthData}
	case setting.AUTH_GOOGLE:
	case setting.AUTH_APPLE:
	}

	// 產生連線Token
	connToken, err := mongo.GenerateConnToken(playerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dbPlayer := &mongo.DBPlayer{
		ID:            playerID,
		CreatedAt:     time.Now(),
		AuthDatas:     authDatas,
		AuthType:      data.AuthType,
		ConnToken:     connToken,
		Gold:          1000,
		Point:         1000,
		OnlineState:   string(setting.Online),
		LastSigninAt:  time.Now(),
		LastSignoutAt: time.Now(),
		Ban:           false,
		DeviceType:    data.DeviceType,
		DeviceUID:     data.DeviceUID,
		InMatchgameID: "",
		MyGladiatorID: "",
	}
	_, err = mongo.AddDocByStruct(mongo.Col.Player, dbPlayer)
	if err != nil {
		errMsg := fmt.Sprintf("%s InitPlayer mongo.AddDocByStruct Player doc 錯誤: %v", logger.LOG_Handler, err)
		handleSignupErr(logData, errMsg)
		http.Error(w, "建立玩家資料失敗", http.StatusInternalServerError)
		return
	}

	logSignup(logData.SetData(dbPlayer.ID, data.AuthType, dbPlayer.OnlineState))

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"data": dbPlayer,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Infof("%s 註冊帳戶 %v 成功: %v", logger.LOG_Handler, playerID, dbPlayer)
}

func handleSignupErr(logData *logger.InitPlayerData, errMsg string) {
	log.Errorf(errMsg)
	logSignup(logData.SetError(errMsg))
}

// logSignup 寫logSignup Log 到 DB
func logSignup(logData *logger.InitPlayerData) {
	mongo.AddDocByStruct(mongo.Col.GameLog, logData)
}
