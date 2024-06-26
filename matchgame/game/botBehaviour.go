package game

import (
	"fmt"
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/utility"
	"matchgame/logger"

	log "github.com/sirupsen/logrus"
)

// 加入Bot
func AddBot() {

	bot, err := GetNewBot()
	if err != nil {
		log.Errorf("%s GetNewBot()錯誤: %v", logger.LOG_BotBehaviour, err)
		return
	}

	err = MyRoom.JoinGamer(bot)
	if err != nil {
		log.Errorf("%s 加入Bot失敗: %v", logger.LOG_BotBehaviour, err)
	}
}

// 建立新Bot玩家
func GetNewBot() (*Bot, error) {
	botIdx := IDAccumulator.GetNextIdx("BotIdx") // 取下一個BotIdx
	botID := fmt.Sprintf("bot%v", botIdx)

	// 取得隨機角鬥士
	rndJsonGladiator, err := gameJson.GetRndJsonGladiator()
	if err != nil {
		return nil, err
	}
	// 隨機角鬥士技能
	normalJsonSkills, err := gameJson.GetJsonSkills("Normal")
	if err != nil {
		return nil, err
	}
	var jsonSkills [GladiatorSkillCount]gameJson.JsonSkill
	rndJsonSkills, err := utility.GetRandomNumberOfTFromMap(normalJsonSkills, 5)
	if err != nil {
		return nil, err
	}
	for i, _ := range rndJsonSkills {
		jsonSkills[i] = rndJsonSkills[i]
	}
	// 設定天賦技能
	talentSkillJson, err := gameJson.GetJsonSkill(rndJsonGladiator.ID)
	if err != nil {
		return nil, err
	}
	jsonSkills[5] = talentSkillJson
	// 設定角鬥士
	gladiator, err := NewGladiator(botID, rndJsonGladiator, jsonSkills, []gameJson.TraitJsonData{}, []gameJson.JsonEquip{})
	if err != nil {
		return nil, err
	}

	// 設定神祉技能
	rndJsonDivineSkills, err := utility.GetRandomNumberOfTFromSlice(MarketDivineJsonSkills[:], DivineSkillCount)
	if err != nil {
		return nil, err
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
	// 設定Bot
	bot := &Bot{
		ID:          botID,
		BribeSkills: divineSkills,
		MyGladiator: &gladiator,
	}
	return bot, nil
}
