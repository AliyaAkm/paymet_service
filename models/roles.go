package models

type Role struct {
	ID   uint   `gorm:"column:id" json:"id"`
	Name string `gorm:"column:name" json:"name"`
	Code string `gorm:"column:code" json:"code"`
}
