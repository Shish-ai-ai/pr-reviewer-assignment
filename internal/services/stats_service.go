package services

import (
	"prReviewerAssignment/internal/db"
	"prReviewerAssignment/internal/models"
	"encoding/json"
	"gorm.io/gorm"
)

type StatsService struct {
	db *gorm.DB
}

func NewStatsService() *StatsService {
	return &StatsService{db: db.DB}
}

func (s *StatsService) GetReviewerStats() (*models.StatsResponse, error) {
	var allPRs []models.PullRequest
	result := s.db.Find(&allPRs)
	if result.Error != nil {
		return nil, result.Error
	}

	reviewerCount := make(map[string]int)
	
	for _, pr := range allPRs {
		var reviewers []string
		if err := json.Unmarshal(pr.AssignedReviewers, &reviewers); err != nil {
			continue
		}
		
		for _, reviewer := range reviewers {
			reviewerCount[reviewer]++
		}
	}

	var stats []models.ReviewerStats
	for userID, count := range reviewerCount {
		stats = append(stats, models.ReviewerStats{
			UserID: userID,
			Count:  count,
		})
	}

	response := &models.StatsResponse{
		ReviewerStats: stats,
	}

	return response, nil
}