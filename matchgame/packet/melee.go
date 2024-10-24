package packet

// logger "matchgame/logger"
// log "github.com/sirupsen/logrus"

type Melee_ToClient struct {
	CMDContent
	MyAttack       PackMelee
	OpponentAttack PackMelee
	MyHandSkillIDs [4]int
}

type PackMelee struct {
	SkillID     int      // 使用技能ID
	MeleePos    float64  // 肉搏位置
	Knockback   float64  // 擊飛強度
	CurPos      float64  // 被擊飛後的位置
	EffectTypes []string // 狀態清單
}
