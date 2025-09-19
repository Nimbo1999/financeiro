package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ID       string
	Email    string
	FullName string
}
