package models

import (
	"time"

	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	ID          string
	Date        time.Time `gorm:"column:t_date"`
	Amount      float64
	Description string `gorm:"column:t_description"`
	UserID      string
}

func (t *Transaction) IsValid() bool {
	return t.UserID != ""
}

func NewTransaction(userID, description string, amount float64, date time.Time) *Transaction {
	return &Transaction{
		UserID:      userID,
		Amount:      amount,
		Date:        date,
		Description: description,
	}
}
