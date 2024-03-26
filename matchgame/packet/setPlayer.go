package packet

import "gladiatorsGoModule/setting"

type SetPlayer struct {
	CMDContent
	DBPlayerID    string
	DBGladiatorID string
}

type SetPlayer_ToClient struct {
	CMDContent
	Players [setting.PLAYER_NUMBER]*PackPlayer
}
type PackPlayer struct {
	DBPlayerID    string
	DBGladiatorID string
}
