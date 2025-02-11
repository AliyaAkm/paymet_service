package models

import (
	"time"
)

type User struct {
	ID                int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name              string    `gorm:"not null" json:"name"`
	Email             string    `gorm:"unique;not null" json:"email"`
	RoleID            uint      `json:"role_id"`
	Password          string    `gorm:"not null" json:"password"`
	IsConfirmed       bool      `json:"-"`
	ConfirmationToken *string   `json:"-"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
