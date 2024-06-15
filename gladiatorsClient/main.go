package main

import (
	// logger "gladiatorsClient/logger"
	"bufio"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"gladiatorsClient/atlas"

	"gladiatorsClient/packet"
	"gladiatorsClient/setting"

	log "github.com/sirupsen/logrus"
	// mongo "gladiatorsGoModule/mongo"
)

type Response struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	DeviceID     string `json:"device_id"`
}

func main() {
	log.Infof("==============gladiatorsClient 啟動==============")

	url := "https://realm.mongodb.com/api/client/v2.0/app/" + setting.AppID + "/auth/providers/anon-user/login"

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	// log.Infof("body= %v", string(body))
	var authRes Response
	if err := json.Unmarshal([]byte(body), &authRes); err != nil {
		log.Fatalf("解析錯誤: %s", err)
	}
	log.Infof("assessToken: %v", authRes.AccessToken)

	atlas.CallAtlasFunc("InitPlayerData", map[string]string{"AuthType": "Guest", "PlayerID": authRes.UserID})
	return
	// 連接到服務器
	conn, err := net.Dial("tcp", "34.81.218.96:7654")
	if err != nil {
		log.Errorf("連線Server錯誤: %v", err)
		os.Exit(1)
	}
	defer conn.Close()
	log.Infof("已連線到Server")

	go func() {
		reader := bufio.NewReader(conn)
		for {
			msg, err := reader.ReadBytes('\n')
			if err != nil {
				log.Errorf("讀取錯誤: %v", err)
				return
			}
			var pack packet.Pack
			if err := json.Unmarshal(msg, &pack); err != nil {
				log.Errorf("解析 Pack 錯誤: %v", err)
				continue
			}
			switch pack.CMD {
			case packet.AUTH_TOCLIENT:
				authToClient, ok := pack.Content.(packet.Auth_ToClient)
				if !ok {
					log.Errorf("Content轉型失敗: %s", packet.AUTH_TOCLIENT)
					return
				}
				log.Infof("Authentication Status: %v, Token: %s", authToClient.IsAuth, authToClient.ConnToken)
			default:
				log.Errorf("未定義的 CMD: %s", pack.CMD)
			}

			log.Infof("收到數據: %s\n", msg)
		}
	}()

	// 送auth封包
	auth := packet.Pack{
		CMD: packet.AUTH,
		Content: packet.Auth{
			Token: authRes.AccessToken,
		},
	}
	packetBytes, err := json.Marshal(auth)
	if err != nil {
		log.Errorf("封包序列化錯誤: %v", err)
		return
	}
	_, err = conn.Write(packetBytes)
	if err != nil {
		log.Errorf("發送錯誤: %v", err)
		return
	}
	log.Info("封包已發送")

}
