package setting

var (
	// 配對類型
	MatchType = MatchTypeStruct{
		Quick: "Quick", // 快速配對
		Test:  "Test",  // 測試房
	}
)

// 配對類型結構
type MatchTypeStruct struct {
	Quick string // 快速配對
	Test  string // 測試房
}

func IsValidMatchType(s string) bool {
	return s == MatchType.Quick || s == MatchType.Test
}
