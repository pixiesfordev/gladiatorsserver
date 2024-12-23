package game

const (
	CONN_TIMEOUT_SECONDS    = 30   // 連線逾時時間
	ROOM_MAX_PLAYER         = 2    // 2個玩家一間房間
	ROOM_MATCH_LOOP_MILISEC = 1000 // 配對每X毫秒循環
	CREATEROOM_WAIT_SECONDS = 1    // 房間建立等待時間
	CREATEROOM_RETRY_TIMES  = 2    // 房間建立嘗試次數
	MATCH_WAIT_MAX_TIME     = 20   // 配對逾時時間
	GAMESERVER_WAIT_TIME    = 30   // 等待遊戲伺服器準備就緒時間
)
