package game

import (
	"fmt"
	"matchgame/logger"

	log "github.com/sirupsen/logrus"
)

// 加入Bot
func AddBot() {
	bot := NewBot()
	err := MyRoom.JoinGamer(bot)
	if err != nil {
		log.Errorf("%s 加入Bot失敗: %v", logger.LOG_BotBehaviour, err)
	}
	err = bot.SetBot()
	if err != nil {
		log := fmt.Sprintf("%s SetBot失敗: %v", logger.LOG_BotBehaviour, err)
		MyRoom.KickBot(bot, log)
	}
}
