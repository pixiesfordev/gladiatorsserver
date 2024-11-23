package game

import (
	"fmt"
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/utility"
	"matchgame/logger"
	"matchgame/packet"
	"time"

	log "github.com/sirupsen/logrus"
)

func (bot *Bot) runBotBehaviour() {
	ticker := time.NewTicker(2 * time.Second) // 每秒判斷一次
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if MyGameState == GAMESTATE_FIGHTING {
				bot.activeRndSkill()
			}
		case <-bot.BehaviourChan.StopChan:
			log.Infof("關閉Bot行為")
			return
		}
	}
}

// 加入Bot
func AddBot() {
	bot := NewBot()
	err := MyRoom.JoinGamer(bot)
	if err != nil {
		log.Errorf("%s 加入Bot失敗: %v", logger.LOG_BotBehaviour, err)
	}
	err = bot.SetBot(bot.ID)
	if err != nil {
		log := fmt.Sprintf("%s SetBot失敗: %v", logger.LOG_BotBehaviour, err)
		MyRoom.KickBot(bot, log)
	}
}

// activeRndSkill 隨機啟用角鬥士技能
func (bot *Bot) activeRndSkill() {
	g := bot.GetGladiator()

	useMelee := utility.GetProbResult(0.5) // 50%機率發動肉搏 50%機率發動立即技能
	if useMelee {
		skills := g.GetAvaliableHandSkillsByActivationType(gameJson.Melee)
		if len(skills) == 0 {
			return
		}
		rndJsonSkill, _ := utility.GetRandomTFromSlice(skills)
		err := g.ActiveSkill(rndJsonSkill, true)
		if err != nil {
			return
		}

	} else {
		skills := g.GetAvaliableHandSkillsByActivationType(gameJson.Instant)
		log.Infof("Vigor: %v    Skills: %v", g.CurVigor, skills)
		if len(skills) == 0 {
			return
		}
		rndJsonSkill, _ := utility.GetRandomTFromSlice(skills)
		err := g.ActiveSkill(rndJsonSkill, true)
		if err != nil {
			return
		}
		myPack := packet.Pack{
			CMD:    packet.PLAYERACTION_TOCLIENT,
			PackID: -1,
			Content: &packet.PlayerAction_ToClient{
				PlayerDBID: bot.GetID(),
				ActionType: packet.INSTANT_SKILL,
				ActionContent: &packet.PackAction_InstantSkill_ToClient{
					SkillID: rndJsonSkill.ID,
				},
			},
		}
		MyRoom.BroadCastPacket(-1, myPack)
	}
}
