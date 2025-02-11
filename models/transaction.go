package models

import "gorm.io/gorm"

type Transaction struct {
	ID             uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	SubscriptionID uint           `json:"subscription_id" gorm:"not null"`                  // Внешний ключ
	Status         string         `json:"status" gorm:"type:varchar(50);default:'pending'"` // pending, paid, declined
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
