package bursary

import (
	"time"

	"github.com/google/uuid"
)

type Ticket struct {
	Id        string                 `json:"id"`
	Channel   string                 `json:"channel"`
	MemberId  string                 `json:"memberId"`
	Expense   int64                  `json:"expense"`
	Income    int64                  `json:"income"`
	Amount    int64                  `json:"amount"` // Amount = Income - Expense
	Fee       int64                  `json:"fee"`
	Total     int64                  `json:"total"` // Amount + Fee
	Desc      string                 `json:"desc"`
	Info      map[string]interface{} `json:"info"`
	CreatedAt time.Time              `json:"createdAt"`
}

func NewTicket() *Ticket {
	return &Ticket{
		Id:        uuid.New().String(),
		Channel:   "default",
		CreatedAt: time.Now(),
	}
}
