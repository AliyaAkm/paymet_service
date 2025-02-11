package models

import "gorm.io/gorm"

type UserSubscription struct {
	ID             uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID         uint           `json:"user_id" gorm:"not null"`
	SubscriptionID uint           `json:"subscription_id" gorm:"not null"`
	StartDate      string         `json:"start_date" gorm:"type:date;not null"` // Дата начала подписки
	EndDate        string         `json:"end_date" gorm:"type:date;not null"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
