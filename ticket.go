package main

import "time"

type Ticket struct {
	LedgerId    string                 `json:"ledgerId"`
	MemberId    string                 `json:"memberId"`
	Rule        string                 `json:"rule"`
	Amount      int                    `json:"amount"`
	Commissions int                    `json:"commission"`
	Total       int                    `json:"total"`
	Desc        string                 `json:"desc"`
	Info        map[string]interface{} `json:"info"`
	CreatedAt   time.Time              `json:"createdAt"`
}
