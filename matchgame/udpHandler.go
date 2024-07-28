package main

import (
	// "gladiatorsGoModule/setting"
	logger "matchgame/logger"
	"sync"

	"encoding/json"

	log "github.com/sirupsen/logrus"

	"matchgame/game"
	"matchgame/packet"
	"net"
	"time"
	// sdk "agones.dev/agones/sdks/go"
)

// 開啟UDP連線
func openConnectUDP(stop chan struct{}, src string) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("%s OpenConnectUDP error: %v.\n", logger.LOG_Main, err)
			stop <- struct{}{}
		}
	}()
	conn, err := net.ListenPacket("udp", src)
	if err != nil {
		log.Errorf("%s (UDP)偵聽失敗: %v.\n", logger.LOG_Main, err)
	}
	defer conn.Close()
	log.Infof("%s (UDP)開始偵聽 %s", logger.LOG_Main, src)

	for {
		// 取得收到的封包
		buffer := make([]byte, 1024)
		n, addr, readBufferErr := conn.ReadFrom(buffer)
		if readBufferErr != nil {
			log.Errorf("%s (UDP)讀取封包錯誤: %v", logger.LOG_Main, readBufferErr)
			continue
		}
		if n <= 0 {
			continue
		}
		// 解析json數據
		var pack packet.UDPReceivePack
		unmarshalErr := json.Unmarshal(buffer[:n], &pack)
		if unmarshalErr != nil {
			log.Errorf("%s (UDP)解析封包錯誤: %s", logger.LOG_Main, unmarshalErr.Error())
			continue
		}
		// 玩家驗證
		player := game.MyRoom.GetPlayerByConnToken(pack.ConnToken)

		if player == nil {
			log.Errorf("%s (UDP)Token驗證失敗 來自 %s 的命令: %s \n", logger.LOG_Main, addr.String(), pack.CMD)
			continue
		}
		if pack.CMD != packet.PING {
			log.Infof("%s (UDP)收到來自 %s 的命令: %s \n", logger.LOG_Main, addr.String(), pack.CMD)
		}

		// 執行命令
		if pack.CMD == packet.UDPAUTH {
			if player.ConnUDP.Conn != nil {
				log.Errorf("%s (UDP)玩家(%s)斷線重連UDP", logger.LOG_Main, player.GetID())
				if player.ConnUDP.Addr.String() != addr.String() { // 玩家通過ConnToken驗證但Addr有變更可能是因為Wifi環境改變
					log.Infof("%s (UDP)玩家 %s 的位置從 %s 變更為 %s \n", logger.LOG_Main, player.GetID(), player.ConnUDP.Addr.String(), addr.String())
				}
			} else {
				// go pingLoop(player, stop)
			}
			// 更新連線資料
			player.ConnUDP.Conn = conn
			player.ConnUDP.Addr = addr
		} else {
			if player.ConnUDP.Conn == nil || player.ConnUDP.Addr == nil {
				log.Warnf("%s (UDP)收到來自 %s(%s) 但尚未進行UDP Auth的命令: %s", logger.LOG_Main, player.GetID(), addr, pack.CMD)
			}
			// 更新連線資料
			player.ConnUDP.Conn = conn

			switch pack.CMD {

			// ==========更新遊戲狀態==========
			case packet.PING:
				// log.Infof("%s 更新玩家 %s 心跳", logger.LOG_Main, player.DBPlayer.ID)
				// player.LastUpdateAt = time.Now() // 更新心跳
			}
		}
	}
}

// 定時更新遊戲狀態給Client
func pingLoop(player *game.Player, stop chan struct{}) {
	log.Infof("%s (UDP)開始updateGameLoop", logger.LOG_Main)
	gameUpdateTimer := time.NewTicker(game.PingMiliSecs * time.Millisecond)

	defer gameUpdateTimer.Stop()

	loopChan := &game.LoopChan{
		StopChan:      make(chan struct{}, 1),
		ChanCloseOnce: sync.Once{},
	}
	player.ConnUDP.MyLoopChan = loopChan

	for {
		select {
		case <-stop:
			log.Infof("強制終止玩家updateGameLoop")
			return
		case <-loopChan.StopChan:
			log.Infof("終止玩家updateGameLoop")
			return
		// ==========心跳==========
		case <-gameUpdateTimer.C:
			pack := packet.Pack{
				CMD:     packet.PING_TOCLIENT,
				PackID:  -1,
				Content: &packet.Ping_ToClient{},
			}
			player.SendPacketToPlayer(pack)
		}

	}
}
