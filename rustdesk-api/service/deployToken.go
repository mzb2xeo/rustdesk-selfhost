package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"rustdesk-api/model"
)

type DeployTokenService struct{}

func (s *DeployTokenService) Create(userId uint, passwordMode, customPassword string) (*model.DeployToken, error) {
	s.CleanupForUser(userId)

	if passwordMode == "" {
		passwordMode = model.DeployPasswordModeStructured
	}

	token, err := s.generateToken()
	if err != nil {
		return nil, err
	}

	dt := &model.DeployToken{
		UserId:         userId,
		Token:          token,
		ExpiresAt:      time.Now().Add(time.Duration(model.DeployTokenTTLSeconds) * time.Second).Unix(),
		PasswordMode:   passwordMode,
		CustomPassword: customPassword,
	}
	if err := DB.Create(dt).Error; err != nil {
		return nil, err
	}
	return dt, nil
}

func (s *DeployTokenService) CleanupForUser(userId uint) {
	now := time.Now().Unix()
	retention := time.Now().AddDate(0, 0, -model.DeployTokenRetentionDays).Unix()
	DB.Where("user_id = ? AND used_at = 0 AND expires_at < ?", userId, now).Delete(&model.DeployToken{})
	DB.Where("user_id = ? AND used_at > 0 AND used_at < ?", userId, retention).Delete(&model.DeployToken{})
}

func (s *DeployTokenService) List(page, pageSize uint, userId uint) *model.DeployTokenList {
	s.CleanupForUser(userId)

	res := &model.DeployTokenList{}
	res.Page = int64(page)
	res.PageSize = int64(pageSize)

	tx := DB.Model(&model.DeployToken{}).Where("user_id = ?", userId)
	tx.Count(&res.Total)
	tx.Order("id desc").Scopes(Paginate(page, pageSize))

	var rows []*model.DeployToken
	tx.Find(&rows)
	for _, row := range rows {
		res.List = append(res.List, row.ToListItem())
	}
	return res
}

func (s *DeployTokenService) InfoByIdForUser(id, userId uint) (*model.DeployToken, error) {
	dt := &model.DeployToken{}
	if err := DB.Where("id = ? AND user_id = ?", id, userId).First(dt).Error; err != nil {
		return nil, errors.New("deploy token not found")
	}
	return dt, nil
}

func (s *DeployTokenService) RevokeById(id, userId uint) error {
	dt, err := s.InfoByIdForUser(id, userId)
	if err != nil {
		return err
	}
	if dt.IsUsed() {
		return errors.New("deploy token already revoked")
	}
	if dt.IsExpired() {
		return errors.New("deploy token already expired")
	}
	now := time.Now().Unix()
	return DB.Model(dt).Update("used_at", now).Error
}

func (s *DeployTokenService) FindValid(token string) (*model.DeployToken, error) {
	dt := &model.DeployToken{}
	if err := DB.Where("token = ?", token).First(dt).Error; err != nil {
		return nil, errors.New("invalid deploy token")
	}
	if dt.IsUsed() {
		return nil, errors.New("deploy token already used")
	}
	if dt.IsExpired() {
		return nil, errors.New("deploy token expired")
	}
	return dt, nil
}

func (s *DeployTokenService) Consume(token string) error {
	dt, err := s.FindValid(token)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	return DB.Model(dt).Update("used_at", now).Error
}

func (s *DeployTokenService) generateToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "rdt_" + hex.EncodeToString(buf), nil
}
