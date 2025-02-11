package models

import "gorm.io/gorm"

type PremiumSubscription struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Plan      string         `json:"plan" gorm:"type:varchar(100);not null"`
	Period    uint           `json:"period" gorm:"not null"`
	Status    string         `json:"status" gorm:"type:varchar(50);default:'active'"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
