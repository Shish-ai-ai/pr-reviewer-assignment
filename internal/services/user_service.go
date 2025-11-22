package services

import (
	"errors"
	"gorm.io/gorm"
	"prReviewerAssignment/internal/db"
	"prReviewerAssignment/internal/models"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService() *UserService {
	return &UserService{db: db.DB}
}

func (s *UserService) SetUserActive(userID string, isActive bool) (*models.User, error) {
	var user models.User
	result := s.db.Where("user_id = ?", userID).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	} else if result.Error != nil {
		return nil, result.Error
	}

	user.IsActive = isActive
	result = s.db.Save(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}
