package game

import "gladiatorsGoModule/utility"

const (
	Init GameState = iota
	Start
	End
)

const (
	TIMELOOP_MILISECS int     = 100 // 遊戲每X毫秒循環
	KICK_PLAYER_SECS  float64 = 20  // 最長允許玩家無心跳X秒後踢出遊戲房
)

var IDAccumulator = utility.NewAccumulator() // 產生一個ID累加器
// Mode模式分為以下:
// standard:一般版本
// non-agones: 個人測試模式(不使用Agones服務, non-agones的連線方式不會透過Matchmaker分配房間再把ip回傳給client, 而是直接讓client去連資料庫matchgame的ip)
var Mode string
