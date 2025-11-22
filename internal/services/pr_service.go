package services

import (
	"encoding/json"
	"errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"math/rand"
	"prReviewerAssignment/internal/db"
	"prReviewerAssignment/internal/models"
	"time"
)

type PRService struct {
	db *gorm.DB
}

func NewPRService() *PRService {
	return &PRService{db: db.DB}
}

func (s *PRService) CreatePullRequest(request models.CreatePRRequest) (*models.PullRequest, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingPR models.PullRequest
	result := tx.Where("pull_request_id = ?", request.PullRequestID).First(&existingPR)
	if result.Error == nil {
		tx.Rollback()
		return nil, errors.New("PR already exists")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, result.Error
	}

	var author models.User
	result = tx.Where("user_id = ? AND is_active = ?", request.AuthorID, true).First(&author)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, errors.New("author not found or inactive")
	} else if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	reviewers, err := s.selectReviewers(tx, author.TeamName, author.UserID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	reviewersJSON, err := json.Marshal(reviewers)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	pr := models.PullRequest{
		PullRequestID:     request.PullRequestID,
		PullRequestName:   request.PullRequestName,
		AuthorID:          request.AuthorID,
		Status:            "OPEN",
		AssignedReviewers: datatypes.JSON(reviewersJSON),
	}

	if err := tx.Create(&pr).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &pr, nil
}

func (s *PRService) selectReviewers(tx *gorm.DB, teamName string, excludeUserID string) ([]string, error) {
	var availableUsers []models.User
	result := tx.Where("team_name = ? AND is_active = ? AND user_id != ?", teamName, true, excludeUserID).Find(&availableUsers)
	if result.Error != nil {
		return nil, result.Error
	}

	if len(availableUsers) == 0 {
		return []string{}, nil
	}

	rand.Shuffle(len(availableUsers), func(i, j int) {
		availableUsers[i], availableUsers[j] = availableUsers[j], availableUsers[i]
	})

	maxReviewers := 2
	if len(availableUsers) < maxReviewers {
		maxReviewers = len(availableUsers)
	}

	var reviewers []string
	for i := 0; i < maxReviewers; i++ {
		reviewers = append(reviewers, availableUsers[i].UserID)
	}

	return reviewers, nil
}

func (s *PRService) MergePullRequest(prID string) (*models.PullRequest, error) {
	var pr models.PullRequest
	result := s.db.Where("pull_request_id = ?", prID).First(&pr)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("PR not found")
	} else if result.Error != nil {
		return nil, result.Error
	}

	if pr.Status == "MERGED" {
		return &pr, nil
	}

	now := time.Now()
	pr.Status = "MERGED"
	pr.MergedAt = &now

	result = s.db.Save(&pr)
	if result.Error != nil {
		return nil, result.Error
	}

	return &pr, nil
}
