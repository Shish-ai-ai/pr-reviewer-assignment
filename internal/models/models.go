package models

import (
	"time"
)

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

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
