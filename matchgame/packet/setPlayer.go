package packet

type SetPlayer struct {
	DBGladiatorID string `json:"DBGladiatorID"`
}

type SetPlayer_ToClient struct {
	Time               int64
	MyPackPlayer       PackPlayer
	OpponentPackPlayer PackPlayer
}
type PackPlayer struct {
	DBID            string
	MyPackGladiator PackGladiator
}

type PackGladiator struct {
	DBID         string   // DBGladiator的DBID
	JsonID       int      // Gladitaor的Json id
	SkillIDs     [6]int   // (玩家自己才會收到)
	HandSkillIDs [4]int   // (玩家自己才會收到)
	MaxHP        int      // 最大生命
	CurHp        int      // 目前生命
	CurVigor     float64  // 目前體力
	CurSpd       float64  // 目前速度
	CurPos       float64  // 目前位置
	EffectTypes  []string // 狀態清單
}
