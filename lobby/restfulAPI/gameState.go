package restfulAPI

import (
	"encoding/json"
	"gladiatorsGoModule/logger"
	"gladiatorsGoModule/mongo"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// [POST] /game/gamestate
func GameState(w http.ResponseWriter, r *http.Request) {
	// 從 header 中獲取 ConnToken
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "GameState取不到 Authorization header", http.StatusUnauthorized)
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		http.Error(w, "錯誤的 Authorization header 格式", http.StatusUnauthorized)
		return
	}
	connToken := tokenParts[1]
	log.Infof("connToken: %v", connToken)

	// 驗證 token
	_, err := mongo.VerifyPlayerByToken(connToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// 獲取遊戲狀態
	gameState, err := mongo.GetDocByID[mongo.DBGameState](mongo.Col.GameSetting, "GameState")
	if err != nil || gameState == nil {
		log.Errorf("%s mongo.GetDocByID 錯誤: %v", logger.LOG_GameState, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 回應結果
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"data": gameState,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
