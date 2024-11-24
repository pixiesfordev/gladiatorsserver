package packet

import ()

// 配對
type Match struct {
	CMDContent
	DBMapID string // DB地圖ID
}

// 配對-送Client
type Match_ToClient struct {
	CMDContent
	CreaterID     string   // 創房者PlayerID
	PlayerIDs     []string // 房間內的所有PlayerID, 索引就是玩家的座位, 一進房間後就不會更動 PlayerIDs[0]就是在座位0玩家的PlayerID
	DBMapID       string   // DB地圖ID
	DbMatchgameID string   // 就是RoomName由Lobby產生，格視為[DBMapID]_[玩家ID]_[時間戳]
	IP            string   // Matchmaker派發Matchgame的IP
	Port          int32    // Matchmaker派發Matchgame的Port
	PodName       string   // Matchmaker派發Matchgame的Pod名稱
}
