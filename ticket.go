package main

import (
	"time"

	"github.com/google/uuid"
)

type Ticket struct {
	ID        string                 `json:"id"`
	Channel   string                 `json:"channel"`
	MemberId  string                 `json:"memberId"`
	Amount    int64                  `json:"amount"`
	Fee       int64                  `json:"fee"`
	Total     int64                  `json:"total"`
	Desc      string                 `json:"desc"`
	Info      map[string]interface{} `json:"info"`
	CreatedAt time.Time              `json:"createdAt"`
}

func NewTicket() *Ticket {
	return &Ticket{
		ID:        uuid.New().String(),
		Channel:   "default",
		CreatedAt: time.Now(),
	}
}
