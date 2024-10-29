package restfulAPI

import (
	"encoding/json"
	"net/http"
	"time"

	"gladiatorsGoModule/logger"
	"gladiatorsGoModule/mongo"

	log "github.com/sirupsen/logrus"
)

// Request 資料
type (
	signinData struct {
		PlayerID   string `json:"playerID"`
		AuthType   string `json:"authType"`
		AuthData   string `json:"authData"`
		DeviceType string `json:"deviceType"`
		DeviceUID  string `json:"deviceUID"`
	}
)

// [POST] /player/signin
func Signin(w http.ResponseWriter, r *http.Request) {
	var data signinData
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dbPlayer, err := mongo.VerifyPlayer(data.AuthType, data.PlayerID, data.AuthData)
	if err != nil || dbPlayer == nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 產生連線Token
	connToken, err := mongo.GenerateConnToken(dbPlayer.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 更新 上線狀態，登入時間，連線Token
	_, err = mongo.UpsertDocByStruct(mongo.Col.Player,
		dbPlayer.ID,
		map[string]interface{}{
			"$set": map[string]interface{}{
				"connToken":    connToken,
				"onlineState":  "Online",
				"lastSigninAt": time.Now(),
			},
		},
	)
	if err != nil {
		log.Errorf("%s UpsertDocByStruct 錯誤: %v", logger.LOG_Signin, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"data": dbPlayer,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
