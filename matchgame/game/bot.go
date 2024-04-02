package game

import (
	gSetting "matchgame/setting"
	"time"
)

// 玩家
type Bot struct {
	ID           string                  // DBot的ID
	MyGladiator  *Gladiator              // 使用中的鬥士
	Gold         int64                   // 玩家金幣
	LastUpdateAt time.Time               // 上次收到玩家更新封包(心跳)
	ConnTCP      *gSetting.ConnectionTCP // TCP連線
	ConnUDP      *gSetting.ConnectionUDP // UDP連線
}

func (bot *Bot) GetID() string {
	return bot.ID
}

func (bot *Bot) GetGold() int64 {
	return bot.Gold
}

func (bot *Bot) AddGold(value int64) {
	bot.Gold += value
}

func (bot *Bot) GetGladiator() *Gladiator {
	return bot.MyGladiator
}
