package logger

import (
	"time"
)

const (
	Init_PlayerData string = "InitPlayerData"
)

type InitPlayerData struct {
	CreatedAt     time.Time  `bson:"createdAt"`
	Type          string     `bson:"type"`
	Error         *string    `bson:"error"`
	PlayerID      *string    `bson:"playerID"`
	AuthType      *string    `bson:"authType"`
	OnlineState   *string    `bson:"onlineState"`
	Point         *int       `bson:"point"`
	LastSigninAt  *time.Time `bson:"lastSigninAt"`
	LastSignoutAt *time.Time `bson:"lastSignoutAt"`
	Ban           *bool      `bson:"ban"`
	DeviceUID     *string    `bson:"deviceUID"`
}

// 建立 初始化玩家資料 Log ORM
func NewSignupData() *InitPlayerData {
	return &InitPlayerData{
		CreatedAt: time.Now(),
		Type:      Init_PlayerData,
	}
}

func (ipd *InitPlayerData) SetError(err string) *InitPlayerData {
	ipd.Error = &err
	return ipd
}

func (ipd *InitPlayerData) SetData(playerID, authType, onlineState string) *InitPlayerData {
	ipd.PlayerID = &playerID
	ipd.AuthType = &authType
	ipd.OnlineState = &onlineState
	return ipd
}

func (ipd *InitPlayerData) SetLastSignTime(lastSigninAt, lastSignoutAt time.Time) *InitPlayerData {
	ipd.LastSigninAt = &lastSigninAt
	ipd.LastSignoutAt = &lastSignoutAt
	return ipd
}

func (ipd *InitPlayerData) SetBan(ban bool) *InitPlayerData {
	ipd.Ban = &ban
	return ipd
}

func (ipd *InitPlayerData) SetPoint(point int) *InitPlayerData {
	ipd.Point = &point
	return ipd
}

func (ipd *InitPlayerData) SetDeviceUID(deviceUID string) *InitPlayerData {
	ipd.DeviceUID = &deviceUID
	return ipd
}
