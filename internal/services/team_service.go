package services

import (
	"errors"
	"gorm.io/gorm"
	"prReviewerAssignment/internal/db"
	"prReviewerAssignment/internal/models"
)

type TeamService struct {
	db *gorm.DB
}

func NewTeamService() *TeamService {
	return &TeamService{db: db.DB}
}

func (s *TeamService) CreateTeam(team models.Team) (*models.Team, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingTeam models.TeamDB
	result := tx.Where("team_name = ?", team.TeamName).First(&existingTeam)
	if result.Error == nil {
		tx.Rollback()
		return nil, errors.New("team already exists")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, result.Error
	}

	teamDB := models.TeamDB{
		TeamName: team.TeamName,
	}
	if err := tx.Create(&teamDB).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, member := range team.Members {
		if err := s.upsertUser(tx, member, team.TeamName); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &team, nil
}

func (s *TeamService) upsertUser(tx *gorm.DB, member models.TeamMember, teamName string) error {
	var existingUser models.User
	result := tx.Where("user_id = ?", member.UserID).First(&existingUser)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		user := models.User{
			UserID:   member.UserID,
			Username: member.Username,
			TeamName: teamName,
			IsActive: member.IsActive,
		}
		return tx.Create(&user).Error
	} else if result.Error == nil {
		return tx.Model(&existingUser).Updates(models.User{
			Username: member.Username,
			TeamName: teamName,
			IsActive: member.IsActive,
		}).Error
	}

	return result.Error
}

func (s *TeamService) GetTeam(teamName string) (*models.Team, error) {
	var teamDB models.TeamDB
	result := s.db.Where("team_name = ?", teamName).First(&teamDB)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("team not found")
	} else if result.Error != nil {
		return nil, result.Error
	}

	var users []models.User
	result = s.db.Where("team_name = ?", teamName).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	var members []models.TeamMember
	for _, user := range users {
		members = append(members, models.TeamMember{
			UserID:   user.UserID,
			Username: user.Username,
			IsActive: user.IsActive,
		})
	}

	team := &models.Team{
		TeamName: teamName,
		Members:  members,
	}

	return team, nil
}
