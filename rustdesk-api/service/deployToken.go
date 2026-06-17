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
	now := time.Now().Unix()
	DB.Where("user_id = ? AND used_at = 0 AND expires_at < ?", userId, now).Delete(&model.DeployToken{})

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
