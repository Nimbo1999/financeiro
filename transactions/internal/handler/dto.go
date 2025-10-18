package handler

import "github.com/nimbo1999/financeiro/transactions/internal/models"

type CreateTransactionRequest struct {
	UserID      string  `json:"user_id"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"` // ISO 8601 format
}

type TransactionVO struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`                 // ISO 8601 format
	CreatedAt   string  `json:"created_at"`           // ISO 8601 format
	UpdatedAt   string  `json:"updated_at"`           // ISO 8601 format
	DeletedAt   string  `json:"deleted_at,omitempty"` // ISO 8601 format
}

func TransactionVOFromModel(tx models.Transaction) TransactionVO {
	deletedAt := ""
	if tx.DeletedAt.Valid {
		deletedAt = tx.DeletedAt.Time.Format("2006-01-02T15:04:05.000Z")
	}
	return TransactionVO{
		ID:          tx.ID,
		UserID:      tx.UserID,
		Description: tx.Description,
		Amount:      tx.Amount,
		Date:        tx.Date.Format("2006-01-02"),
		CreatedAt:   tx.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		UpdatedAt:   tx.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
		DeletedAt:   deletedAt,
	}
}
