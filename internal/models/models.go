package models

import (
	"gorm.io/datatypes"
	"time"
)

type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type User struct {
	UserID    string    `gorm:"primaryKey" json:"user_id"`
	Username  string    `gorm:"not null" json:"username"`
	TeamName  string    `gorm:"not null" json:"team_name"`
	IsActive  bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"-"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"-"`
}

type TeamDB struct {
	TeamName  string    `gorm:"primaryKey" json:"team_name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"-"`
}

type PullRequest struct {
	PullRequestID     string         `gorm:"primaryKey" json:"pull_request_id"`
	PullRequestName   string         `gorm:"not null" json:"pull_request_name"`
	AuthorID          string         `gorm:"not null" json:"author_id"`
	Status            string         `gorm:"type:varchar(20);not null;default:'OPEN';check:status IN ('OPEN', 'MERGED')" json:"status"`
	AssignedReviewers datatypes.JSON `gorm:"type:jsonb" json:"assigned_reviewers"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	MergedAt          *time.Time     `json:"mergedAt,omitempty"`

	Author User `gorm:"foreignKey:AuthorID;references:UserID" json:"-"`
}

type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}
