package setting

const (
	// 命名空間
	NAMESPACE_MATCHERSERVER = "gladiators-service"    // 配對伺服器命名空間
	NAMESPACE_GAMESERVER    = "gladiators-gameserver" // 遊戲伺服器命名空間

	// 服務名稱
	MATCHMAKER            = "gladiators-matchmaker"    // 配對伺服器Services名稱
	MATCHGAME_TESTVER_TCP = "gladiators-matchgame-tcp" // 個人測試用遊戲房Services TCP名稱
	MATCHGAME_TESTVER_UDP = "gladiators-matchgame-udp" // 個人測試用遊戲房Services UDP名稱

	// 遊戲房舍定
	PLAYER_NUMBER = 2 // 遊戲房最多X位玩家
)

var EnvGCPProject = map[string]string{
	"Dev":     "mygladiators-dev",   // 開發版
	"Release": "mygladiators-release", // 正式版
}

// 環境版本
const (
	ENV_DEV     = "Dev"
	ENV_RELEASE = "Release"
)

// 配對類型結構
type MatchTypeStruct struct {
	Quick string // 快速配對
	Test  string // 測試房
}

// 配對類型
var MatchType = MatchTypeStruct{
	Quick: "Quick", // 快速配對
	Test:  "Test",  // 測試房
}
