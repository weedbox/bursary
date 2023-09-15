package bursary

import (
	"time"

	"github.com/google/uuid"
)

type Ticket struct {
	ID        string                 `json:"id"`
	Channel   string                 `json:"channel"`
	MemberID  string                 `json:"member_id"`
	Expense   int64                  `json:"expense"`
	Income    int64                  `json:"income"`
	Amount    int64                  `json:"amount"` // Amount = Income - Expense
	Fee       int64                  `json:"fee"`
	Total     int64                  `json:"total"` // Amount + Fee
	Desc      string                 `json:"desc"`
	Info      map[string]interface{} `json:"info"`
	CreatedAt time.Time              `json:"created_at"`
}

func NewTicket() *Ticket {
	return &Ticket{
		ID:        uuid.New().String(),
		Channel:   "default",
		CreatedAt: time.Now(),
	}
}
