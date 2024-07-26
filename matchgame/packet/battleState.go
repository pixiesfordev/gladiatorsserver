package packet

// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"

type BattleState struct {
	CMDContent
}

type BattleState_ToClient struct {
	CMDContent
	MyPlayerState       PackPlayerState
	OpponentPlayerState PackPlayerState
	GameTime            float64
}

type PackPlayerState struct {
	DBID           string // 玩家DBID
	DivineSkills   [2]PackDivineSkill
	GladiatorState PackGladiatorState
}

type PackDivineSkill struct {
	JsonID int
	Used   bool
}

type PackGladiatorState struct {
	HandSkillIDs            [4]int   // (玩家自己才會收到)
	CurHp                   int      // 目前生命
	CurVigor                float64  // 目前體力
	CurSpd                  float64  // 目前速度
	CurPos                  float64  // 目前位置
	IsRush                  bool     // 是否正在衝刺中
	EffectTypes             []string // 狀態清單
	ActivedMeleeJsonSkillID int      // (玩家自己才會收到)啟用中的肉搏技能ID, 玩家啟用中的肉搏技能, 如果是0代表沒有啟用中的肉搏技能
}
