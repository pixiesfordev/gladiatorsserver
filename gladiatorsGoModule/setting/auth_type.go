package setting

// 註冊類型
type AuthType string

const (
	Guest    AuthType = "Guest"    // 訪客
	Official AuthType = "Official" // 官方註冊
	Unknown  AuthType = "Unknown"  // 未知錯誤
)

// 檢查合法 AuthType
func ParseAuthType(s string) (AuthType, bool) {
	switch AuthType(s) {
	case Guest, Official, Unknown:
		return AuthType(s), true
	}

	return "", false
}
