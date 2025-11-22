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

func (s *PRService) ReassignReviewer(prID string, oldReviewerID string) (*models.PullRequest, string, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, "", tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var pr models.PullRequest
	result := tx.Where("pull_request_id = ?", prID).First(&pr)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, "", errors.New("PR not found")
	} else if result.Error != nil {
		tx.Rollback()
		return nil, "", result.Error
	}

	if pr.Status == "MERGED" {
		tx.Rollback()
		return nil, "", errors.New("cannot reassign on merged PR")
	}

	var reviewers []string
	if err := json.Unmarshal(pr.AssignedReviewers, &reviewers); err != nil {
		tx.Rollback()
		return nil, "", err
	}

	found := false
	for _, reviewer := range reviewers {
		if reviewer == oldReviewerID {
			found = true
			break
		}
	}
	if !found {
		tx.Rollback()
		return nil, "", errors.New("reviewer is not assigned to this PR")
	}

	var oldReviewer models.User
	result = tx.Where("user_id = ? AND is_active = ?", oldReviewerID, true).First(&oldReviewer)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, "", errors.New("old reviewer not found or inactive")
	} else if result.Error != nil {
		tx.Rollback()
		return nil, "", result.Error
	}

	newReviewer, err := s.findReplacementCandidate(tx, oldReviewer.TeamName, pr.AuthorID, reviewers)
	if err != nil {
		tx.Rollback()
		return nil, "", err
	}

	for i, reviewer := range reviewers {
		if reviewer == oldReviewerID {
			reviewers[i] = newReviewer
			break
		}
	}

	reviewersJSON, err := json.Marshal(reviewers)
	if err != nil {
		tx.Rollback()
		return nil, "", err
	}
	pr.AssignedReviewers = datatypes.JSON(reviewersJSON)

	result = tx.Save(&pr)
	if result.Error != nil {
		tx.Rollback()
		return nil, "", result.Error
	}

	if err := tx.Commit().Error; err != nil {
		return nil, "", err
	}

	return &pr, newReviewer, nil
}

func (s *PRService) findReplacementCandidate(tx *gorm.DB, teamName string, authorID string, currentReviewers []string) (string, error) {
	var availableUsers []models.User

	query := "team_name = ? AND is_active = ? AND user_id != ?"
	params := []interface{}{teamName, true, authorID}

	if len(currentReviewers) > 0 {
		query += " AND user_id NOT IN (?"
		params = append(params, currentReviewers[0])
		for i := 1; i < len(currentReviewers); i++ {
			query += ",?"
			params = append(params, currentReviewers[i])
		}
		query += ")"
	}

	result := tx.Where(query, params...).Find(&availableUsers)
	if result.Error != nil {
		return "", result.Error
	}

	if len(availableUsers) == 0 {
		return "", errors.New("no active replacement candidate in team")
	}

	rand.Shuffle(len(availableUsers), func(i, j int) {
		availableUsers[i], availableUsers[j] = availableUsers[j], availableUsers[i]
	})

	return availableUsers[0].UserID, nil
}
