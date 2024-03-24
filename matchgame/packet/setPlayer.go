package packet

type SetPlayer struct {
	CMDContent
	DBPlayerID    string
	DBGladiatorID string
}

type SetPlayer_ToClient struct {
	CMDContent
	Players [4]*PackPlayer
}
type PackPlayer struct {
	DBPlayerID    string
	DBGladiatorID string
}
