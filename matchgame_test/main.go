package main

import (
	"bufio"
	"fmt"
	"gladiatorsGoModule/mongo"
	"net"
	"os"

	log "github.com/sirupsen/logrus"
)

var conn net.Conn

func main() {

	mongo.Init(mongo.InitData{
		Env:           "Dev",
		APIPublicKey:  "faakhiny",
		APIPrivateKey: "f751e44f-d215-49ac-9883-a30f0f1db397",
	}, "pixiesfordev", "qbTDhfmeItCF82Gr")
	// 建立連線
	gameState, err := mongo.GetDocByID[mongo.DBGameState](mongo.Col.GameSetting, "GameState")
	if err != nil {
		log.Errorf("取gameState失敗: %v", err)
		return
	}
	address := fmt.Sprintf("%v:%v", gameState.MatchgameTestverTcpIP, gameState.MatchgameTestverPort)
	log.Infof(" Address: %v", address)
	conn, err = net.Dial("tcp", address)
	if err != nil {
		log.Infof("連線到伺服器時發生錯誤: %v", err)
		os.Exit(1)
	}
	defer conn.Close()

	log.Infof("已連線到Server")

	initKeyboard()

	// 送出AUTH封包
	send_Auth()

	// 接收並處理來自伺服器的訊息
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		data := scanner.Text()
		processData(data)
	}

	if err := scanner.Err(); err != nil {
		log.Infof("讀取資料錯誤: %v", err)
	}
}
