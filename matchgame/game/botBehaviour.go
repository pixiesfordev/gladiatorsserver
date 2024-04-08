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
		return nil, fmt.Errorf("gameJson.GetRndJsonGladiator()錯誤: %v", err)
	}
	// 隨機角鬥士技能
	allJsonSkills, err := gameJson.GetJsonSkills()
	if err != nil {
		return nil, fmt.Errorf("gameJson.GetJsonSkills()錯誤: %v", err)
	}
	var jsonSkills [6]gameJson.JsonSkill
	rndJsonSkills, err := utility.GetRandomNumberOfTFromMap(allJsonSkills, 5)
	if err != nil {
		return nil, fmt.Errorf("utility.GetRandomNumberOfTFromMap錯誤: %v", err)
	}
	for i, _ := range rndJsonSkills {
		jsonSkills[i] = rndJsonSkills[i]
	}
	// 設定天賦技能
	talentSkillJson, err := gameJson.GetJsonSkill(rndJsonGladiator.ID)
	if err != nil {
		return nil, fmt.Errorf("gameJson.GetJsonSkill(rndJsonGladiator.ID)錯誤: %v", err)
	}
	jsonSkills[5] = talentSkillJson
	// 設定角鬥士
	gladiator := Gladiator{
		ID:            botID,
		JsonGladiator: rndJsonGladiator,
		JsonSkills:    jsonSkills,
	}
	// 設定Bot
	bot := &Bot{
		ID:          botID,
		MyGladiator: &gladiator,
	}
	return bot, nil
}
