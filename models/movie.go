package models

import "gorm.io/gorm"

type Movie struct {
	gorm.Model
	Title       string
	Description string
	Price       float64
	Genre       string
	ReleaseDate string
	ImageURL    string
}
