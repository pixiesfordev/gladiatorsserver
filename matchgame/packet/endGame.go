package packet

import (
// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"
)

type EndGame_ToClient struct {
	CMDContent
	Result        string // Die(有一方死亡), Surrender(有一方投降), Timeout(時間到)
	PlayerResults []PackPlayerResult
}

type PackPlayerResult struct {
	Result     string // Win,Lose,Tie
	DBPlayerID string
	GainGold   int
	GainEXP    int
}
