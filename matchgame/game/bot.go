package game

import (
	"fmt"
	"gladiatorsGoModule/setting"
	"matchgame/packet"
)

// 玩家
type Bot struct {
	ID          string                         // DBot的ID
	Idx         int                            // 第一位玩家是0(左方) 第二位玩家是1(右方)
	MyGladiator *Gladiator                     // 使用中的鬥士
	BribeSkills [DivineSkillCount]*DivineSkill // 神祉技能
}

func NewBot(gladiator *Gladiator, divineSkills [DivineSkillCount]*DivineSkill) *Bot {
	botIdx := IDAccumulator.GetNextIdx("BotIdx") // 取下一個BotIdx
	botID := fmt.Sprintf("bot%v", botIdx)
	bot := &Bot{
		ID:          botID,
		BribeSkills: divineSkills,
		MyGladiator: gladiator,
	}
	return bot
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
func (bot *Bot) Surrender() {

}
func (bot *Bot) IsSelectedDivineSkill() bool {
	return true
}
func (bot *Bot) GetPackPlayerBribes() [setting.PLAYER_NUMBER]packet.PackDivineSkill {
	var botBribes [2]packet.PackDivineSkill

	botBribes[0] = packet.PackDivineSkill{
		JsonID: bot.BribeSkills[0].MyJson.ID,
		Used:   bot.BribeSkills[0].Used,
	}
	botBribes[1] = packet.PackDivineSkill{
		JsonID: bot.BribeSkills[1].MyJson.ID,
		Used:   bot.BribeSkills[1].Used,
	}

	return botBribes
}

// GetPackPlayer 取得玩家封包
func (bot *Bot) GetPackPlayer(myself bool) packet.PackPlayer {
	packPlayer := packet.PackPlayer{
		DBID:            bot.GetID(),
		MyPackGladiator: bot.MyGladiator.GetPackGladiator(myself),
	}
	return packPlayer
}

func (bot *Bot) GetPackPlayerState(myselfPack bool) packet.PackPlayerState {
	packBotState := packet.PackPlayerState{
		DBID:           bot.GetID(),
		DivineSkills:   bot.GetPackPlayerBribes(),
		GladiatorState: bot.GetGladiator().GetPackGladiatorState(myselfPack),
	}
	return packBotState
}

func (bot *Bot) GetPackDivineSkills() [setting.PLAYER_NUMBER]packet.PackDivineSkill {
	var packDivineSkills [2]packet.PackDivineSkill
	return packDivineSkills
}
