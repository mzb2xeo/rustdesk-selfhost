package model

import "time"

const DeployTokenTTLSeconds = 30 * 60

// DeployTokenRetentionDays removes used tokens older than this from the database.
const DeployTokenRetentionDays = 7

const (
	DeployTokenStatusActive  = "active"
	DeployTokenStatusUsed    = "used"
	DeployTokenStatusExpired = "expired"
)

const (
	DeployPasswordModeStructured = "structured"
	DeployPasswordModeCustom     = "custom"
)

type DeployToken struct {
	IdModel
	UserId         uint   `json:"user_id" gorm:"default:0;not null;index"`
	Token          string `json:"token" gorm:"default:'';not null;uniqueIndex"`
	ExpiresAt      int64  `json:"expires_at" gorm:"default:0;not null;index"`
	UsedAt         int64  `json:"used_at" gorm:"default:0;not null;"`
	PasswordMode   string `json:"password_mode" gorm:"default:'structured';not null;"`
	CustomPassword string `json:"-" gorm:"default:'';not null;"`
	TimeModel
}

func (dt *DeployToken) IsExpired() bool {
	return dt.ExpiresAt > 0 && dt.ExpiresAt < time.Now().Unix()
}

func (dt *DeployToken) IsUsed() bool {
	return dt.UsedAt > 0
}

func (dt *DeployToken) Status() string {
	if dt.IsUsed() {
		return DeployTokenStatusUsed
	}
	if dt.IsExpired() {
		return DeployTokenStatusExpired
	}
	return DeployTokenStatusActive
}

func (dt *DeployToken) TokenPreview() string {
	t := dt.Token
	if len(t) <= 16 {
		return t
	}
	return t[:10] + "..." + t[len(t)-6:]
}

type DeployTokenListItem struct {
	Id           uint   `json:"id"`
	TokenPreview string `json:"token_preview"`
	Status       string `json:"status"`
	PasswordMode string `json:"password_mode"`
	ExpiresAt    int64  `json:"expires_at"`
	UsedAt       int64  `json:"used_at"`
	CreatedAt    int64  `json:"created_at"`
}

func (dt *DeployToken) ToListItem() *DeployTokenListItem {
	return &DeployTokenListItem{
		Id:           dt.Id,
		TokenPreview: dt.TokenPreview(),
		Status:       dt.Status(),
		PasswordMode: dt.PasswordMode,
		ExpiresAt:    dt.ExpiresAt,
		UsedAt:       dt.UsedAt,
		CreatedAt:    time.Time(dt.CreatedAt).Unix(),
	}
}

type DeployTokenList struct {
	List []*DeployTokenListItem `json:"list,omitempty"`
	Pagination
}
