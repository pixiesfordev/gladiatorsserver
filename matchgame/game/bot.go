package game

import ()

// 玩家
type Bot struct {
	ID          string     // DBot的ID
	MyGladiator *Gladiator // 使用中的鬥士
}

func (bot *Bot) GetID() string {
	return bot.ID
}

func (bot *Bot) GetGold() int64 {
	return 0
}

func (bot *Bot) AddGold(value int64) {
}

func (bot *Bot) GetGladiator() *Gladiator {
	return bot.MyGladiator
}

func (bot *Bot) IsReady() bool {
	return true
}
