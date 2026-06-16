package model

import "time"

const DeployTokenTTLSeconds = 30 * 60

type DeployToken struct {
	IdModel
	UserId    uint   `json:"user_id" gorm:"default:0;not null;index"`
	Token     string `json:"token" gorm:"default:'';not null;uniqueIndex"`
	ExpiresAt int64  `json:"expires_at" gorm:"default:0;not null;index"`
	UsedAt    int64  `json:"used_at" gorm:"default:0;not null;"`
	TimeModel
}

func (dt *DeployToken) IsExpired() bool {
	return dt.ExpiresAt > 0 && dt.ExpiresAt < time.Now().Unix()
}

func (dt *DeployToken) IsUsed() bool {
	return dt.UsedAt > 0
}
