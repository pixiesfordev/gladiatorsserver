package game

import (
	"fmt"
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/utility"
	"matchgame/logger"

	log "github.com/sirupsen/logrus"
)

// 加入Bot
func AddBot() *Bot {
	botIdx := IDAccumulator.GetNextIdx("BotIdx") // 取下一個BotIdx
	botID := fmt.Sprintf("bot%v", botIdx)

	// 取得隨機角鬥士
	rndJsonGladiator, err := gameJson.GetRndJsonGladiator()
	if err != nil {
		log.Errorf("%s gameJson.GetRndJsonGladiator()錯誤: %v", logger.LOG_BotBehaviour, err)
		return nil
	}

	allJsonSkills, err := gameJson.GetJsonSkills()
	if err != nil {
		log.Errorf("%s gameJson.GetJsonSkills()錯誤: %v", logger.LOG_BotBehaviour, err)
		return nil
	}

	rndJsonSkills, err := utility.GetRandomNumberOfTFromMap(allJsonSkills, 5)
	if err != nil {
		log.Errorf("%s utility.GetRandomNumberOfTFromMap錯誤: %v", logger.LOG_BotBehaviour, err)
		return nil
	}

	var jsonSkills [6]gameJson.JsonSkill

	gladiator := Gladiator{
		ID:            botID,
		JsonGladiator: rndJsonGladiator,
		JsonSkills:    jsonSkills,
	}

	// hero := Hero{
	// 	ID:     int(heroID),
	// 	skinID: heroSkinID,
	// 	spells: skillJsons,
	// }

	// bot := Bot{
	// 	ID:           botID,
	// 	MyHero:       &hero,
	// 	curTargetIdx: -1, // 無攻擊目標時, curTargetIdx為-1
	// }
	// bot.InitHero(0, [3]int{}, [3]float64{})
	// joined := MyRoom.JoinPlayer(&bot)
	// if !joined {
	// 	log.Errorf("%s 玩家加入房間失敗", logger.LOG_Main)
	// 	return nil
	// }

	// // 廣播更新玩家
	// MyRoom.BroadCastPacket(-1, &packet.Pack{
	// 	CMD:    packet.UPDATEPLAYER_TOCLIENT,
	// 	PackID: -1,
	// 	Content: &packet.UpdatePlayer_ToClient{
	// 		Players: MyRoom.GetPacketPlayers(),
	// 	},
	// })
	// // 廣播角鬥士選擇
	// heroIDs, heroSkinIDs := MyRoom.GetHeroInfos()
	// MyRoom.BroadCastPacket(-1, &packet.Pack{
	// 	CMD: packet.SETHERO_TOCLIENT,
	// 	Content: &packet.SetHero_ToClient{
	// 		HeroIDs:     heroIDs,
	// 		HeroSkinIDs: heroSkinIDs,
	// 	},
	// })
	// bot.newSelectTargetLoop()
	// return &bot
	return nil
}
