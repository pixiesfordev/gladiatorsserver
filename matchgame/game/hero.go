package game

import (
	"fmt"
	"gladiatorsGoModule/gameJson"

	log "github.com/sirupsen/logrus"

	// "gladiatorsGoModule/utility"
	"matchgame/logger"
)

type Hero struct {
	ID     int                           // 英雄ID
	SkinID string                        // SkinID
	EXP    int                           // 英雄經驗
	Spells [3]gameJson.HeroSpellJsonData // 英雄技能
}

// 取得英雄技能
func (hero *Hero) GetSpell(idx int) (gameJson.HeroSpellJsonData, error) {
	if idx < 1 || idx > 3 {
		log.Errorf("%s GetSpell傳入錯誤索引: %v", logger.LOG_Setting, idx)
		return gameJson.HeroSpellJsonData{}, fmt.Errorf("%s GetSpell傳入錯誤索引: %v", logger.LOG_Setting, idx)
	}
	return hero.Spells[(idx - 1)], nil // Spells索引是存0~2所以idx要-1
}
