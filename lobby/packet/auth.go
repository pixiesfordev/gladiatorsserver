package packet

// 帳號登入
type Auth struct {
	CMDContent
	ConnToken string
}

// 帳號登入回傳client
type Auth_ToClient struct {
	CMDContent
	IsAuth bool // 是否驗證成功
	Time   int64
}
