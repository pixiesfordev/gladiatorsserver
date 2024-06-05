package game

import (
	"gladiatorsGoModule/setting"
	"matchgame/packet"
)

// 玩家
type Bot struct {
	ID          string                       // DBot的ID
	Idx         int                          // 第一位玩家是0(左方) 第二位玩家是1(右方)
	MyGladiator *Gladiator                   // 使用中的鬥士
	BribeSkills [BribeSkillCount]*BribeSkill // 賄賂技能
}

func (bot *Bot) SetIdx(idx int) {
	bot.Idx = idx
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
func (bot *Bot) GetPackPlayerBribes() [setting.PLAYER_NUMBER]packet.PackBribeSkill {
	var botBribes [2]packet.PackBribeSkill

	botBribes[0] = packet.PackBribeSkill{
		JsonID: bot.BribeSkills[0].MyJson.ID,
		Used:   bot.BribeSkills[0].Used,
	}
	botBribes[1] = packet.PackBribeSkill{
		JsonID: bot.BribeSkills[1].MyJson.ID,
		Used:   bot.BribeSkills[1].Used,
	}

	return botBribes
}

func (bot *Bot) GetPackPlayerState() packet.PackPlayerState {
	packBotState := packet.PackPlayerState{
		ID:          bot.GetID(),
		BribeSkills: bot.GetPackPlayerBribes(),
		Gladiator:   bot.GetGladiator().GetPackGladiator(),
	}
	return packBotState
}
