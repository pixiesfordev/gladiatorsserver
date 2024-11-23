package packet

type GameState_ToClient struct {
	CMDContent
	State string // 參考game.go的GameState
}

type Hp_ToClient struct {
	CMDContent
	PlayerID   string
	HPChange   int
	EffectType string
	CurHp      int
	MaxHp      int
}

// 角鬥士狀態
type GladiatorStates_ToClient struct {
	CMDContent
	Time          int64
	MyState       PackGladiatorState
	OpponentState PackGladiatorState
}
type PackGladiatorState struct {
	CurPos      float64      // 目前位置
	CurSpd      float64      // 目前速度
	CurVigor    float64      // 目前體力
	Rush        bool         // 是否衝刺中
	EffectDatas []PackEffect // Buff狀態資料
}
type PackEffect struct {
	EffectName string  // 特效名稱
	Duration   float64 // 特效時間
}
