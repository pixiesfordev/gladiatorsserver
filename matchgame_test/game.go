package main

import (
	log "github.com/sirupsen/logrus"
	"matchgame/game"
	"time"
)

func changeGameState(gameState game.GameState) {
	switch gameState {
	case game.GAMESTATE_WAITINGPLAYERS:
		go startPingLoop()
	}
}

func startPingLoop() {
	log.Infof("開始PING LOOP")
	ticker := time.NewTicker(1 * time.Second) // 每秒觸發一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			send_Ping()
		}
	}
}
