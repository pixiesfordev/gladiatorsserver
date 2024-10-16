package game

import (
	"fmt"
	"gladiatorsGoModule/gameJson"
	// "gladiatorsGoModule/setting"
	"gladiatorsGoModule/utility"
	"matchgame/packet"
)

// 玩家
type Bot struct {
	ID          string                         // DBot的ID
	Idx         int                            // 第一位玩家是0(左方) 第二位玩家是1(右方)
	opponent    Gamer                          // 對手
	MyGladiator *Gladiator                     // 使用中的鬥士
	BribeSkills [DivineSkillCount]*DivineSkill // 神祉技能
}

func NewBot() *Bot {
	botIdx := IDAccumulator.GetNextIdx() // 取下一個BotIdx
	botID := fmt.Sprintf("bot%v", botIdx)
	bot := &Bot{
		ID: botID,
	}
	return bot
}

func (bot *Bot) SetBot(botID string) error {

	// 取得隨機角鬥士
	rndJsonGladiator, err := gameJson.GetRndJsonGladiator()
	if err != nil {
		return err
	}
	// 隨機角鬥士技能
	normalJsonSkills, err := gameJson.GetJsonSkills("Normal")
	if err != nil {
		return err
	}
	var jsonSkills [GladiatorSkillCount]gameJson.JsonSkill
	rndJsonSkills, err := utility.GetRandomNumberOfTFromMap(normalJsonSkills, 5)
	if err != nil {
		return err
	}
	for i, _ := range rndJsonSkills {
		jsonSkills[i] = rndJsonSkills[i]
	}
	// 設定天賦技能
	talentSkillJson, err := gameJson.GetJsonSkill(rndJsonGladiator.ID)
	if err != nil {
		return err
	}
	jsonSkills[5] = talentSkillJson
	// 設定角鬥士
	gladiator, err := NewGladiator(bot, "botGladiator", rndJsonGladiator, jsonSkills, []gameJson.TraitJson{}, []gameJson.JsonEquip{})
	if err != nil {
		return err
	}

	// 設定神祉技能
	rndJsonDivineSkills, err := utility.GetRandomNumberOfTFromSlice(MarketDivineJsonSkills[:], DivineSkillCount)
	if err != nil {
		return err
	}
	var divineSkills [DivineSkillCount]*DivineSkill
	for i, _ := range divineSkills {
		if i < len(rndJsonDivineSkills) {
			divineSkills[i] = &DivineSkill{
				Used:   false,
				MyJson: rndJsonDivineSkills[i],
			}
		} else {
			divineSkills[i] = nil
		}
	}

	bot.MyGladiator = &gladiator
	bot.BribeSkills = divineSkills

	return nil
}

func (bot *Bot) SetIdx(idx int) {
	bot.Idx = idx
}
func (bot *Bot) GetID() string {
	return bot.ID
}

// GetOpponent 取得對手Gamer
func (bot *Bot) GetOpponent() Gamer {
	return bot.opponent
}

// SetOpponent 設定對手Gamer
func (bot *Bot) SetOpponent(gamer Gamer) {
	bot.opponent = gamer
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

// GetPackPlayer 取得玩家封包
func (bot *Bot) GetPackPlayer(myself bool) packet.PackPlayer {
	packPlayer := packet.PackPlayer{
		DBID:            bot.GetID(),
		MyPackGladiator: bot.MyGladiator.GetPackGladiator(myself),
	}
	return packPlayer
}
