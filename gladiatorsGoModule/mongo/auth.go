package mongo

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"gladiatorsGoModule/logger"
	"gladiatorsGoModule/setting"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func VerifyPlayer(authType, playerID, authData string) (*DBPlayer, error) {
	var dbPlayer *DBPlayer
	err := GetDocByID(Col.Player, playerID, dbPlayer)
	if err != nil || dbPlayer == nil {
		return nil, fmt.Errorf("%v VerifyPlayer玩家資料錯誤: %v", logger.LOG_Mongo, err)
	}

	verify := false
	switch authType {
	case setting.AUTH_GUSET:
		verify = authData == dbPlayer.AuthDatas[setting.AUTH_GUSET]
	default:
		return nil, fmt.Errorf("%v VerifyPlayer驗證類型錯誤: %v", logger.LOG_Mongo, err)
	}

	if !verify {
		return nil, fmt.Errorf("%v VerifyPlayer驗證失敗: %v", logger.LOG_Mongo, err)
	}

	return dbPlayer, nil
}

// GenerateConnToken 產生socket連線用token
func GenerateConnToken(playerID string) (string, error) {
	if playerID == "" {
		return "", fmt.Errorf("%v NewConnToken傳入的playerID為空", logger.LOG_Mongo)
	}
	timestamp := time.Now().UnixNano()
	data := fmt.Sprintf("%s:%d", playerID, timestamp)
	hash := sha256.New()
	hash.Write([]byte(data))
	token := hex.EncodeToString(hash.Sum(nil))

	return token, nil
}

func VerifyPlayerByToken(connToken string) (*DBPlayer, error) {
	dbPlayer, err := GetDocByFilter[DBPlayer](Col.Player, bson.M{"connToken": connToken})
	if err != nil || dbPlayer == nil {
		return nil, err
	}
	return dbPlayer, nil
}
