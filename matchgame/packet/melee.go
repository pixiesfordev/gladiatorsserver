package packet

// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"

type PackMelee struct {
	SkillID     int          // 使用技能ID
	MeleePos    float64      // 肉搏位置
	Knockback   float64      // 擊飛強度
	CurPos      float64      // 被擊飛後的位置
	EffectDatas []PackEffect // Buff狀態資料
}

type Melee_ToClient struct {
	CMDContent
	MyAttack       PackMelee
	OpponentAttack PackMelee
	SkillOnID      int    // 啟用中的肉搏技能
	NewSkilID      int    // 新抽到的技能
	MyHandSkillIDs [4]int // 目前手牌
}
type BeforeMeleeSkill_ToClient struct {
	MySkillID       int
	OpponentSkillID int
}
