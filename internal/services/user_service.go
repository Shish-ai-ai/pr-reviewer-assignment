package services

import (
	"encoding/json"
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

func (s *UserService) GetUserReviews(userID string) (*models.UserReviewResponse, error) {
	var user models.User
	result := s.db.Where("user_id = ?", userID).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	} else if result.Error != nil {
		return nil, result.Error
	}

	var allPRs []models.PullRequest
	result = s.db.Find(&allPRs)
	if result.Error != nil {
		return nil, result.Error
	}

	var userPRs []models.PullRequestShort
	for _, pr := range allPRs {
		var reviewers []string
		if err := json.Unmarshal(pr.AssignedReviewers, &reviewers); err != nil {
			continue
		}

		for _, reviewer := range reviewers {
			if reviewer == userID {
				userPRs = append(userPRs, models.PullRequestShort{
					PullRequestID:   pr.PullRequestID,
					PullRequestName: pr.PullRequestName,
					AuthorID:        pr.AuthorID,
					Status:          pr.Status,
				})
				break
			}
		}
	}

	response := &models.UserReviewResponse{
		UserID:       userID,
		PullRequests: userPRs,
	}

	return response, nil
}
