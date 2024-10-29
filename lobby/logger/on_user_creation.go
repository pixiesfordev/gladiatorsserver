package logger

import (
	"time"
)

const (
	On_UserCreation string = "OnUserCreation"
)

type UserCreation struct {
	CreatedAt time.Time `bson:"createdAt"`
	Type      string    `bson:"type"`
	Error     *string   `bson:"error"`
	PlayerID  *string   `bson:"playerID"`
	Role      *string   `bson:"role"`
}

// 建立 玩家建立資料 Log ORM
func NewUserCreation() *UserCreation {
	return &UserCreation{
		CreatedAt: time.Now(),
		Type:      On_UserCreation,
	}
}

func (uc *UserCreation) SetError(err string) *UserCreation {
	uc.Error = &err
	return uc
}

func (uc *UserCreation) SetPlayerID(playerID string) *UserCreation {
	uc.PlayerID = &playerID
	return uc
}

func (uc *UserCreation) SetRole(playerID string) *UserCreation {
	uc.PlayerID = &playerID
	return uc
}
