package main

import "time"

type LedgerEntry struct {
	MemberId    string                 `json:"memberId"`
	Amount      int                    `json:"amount"`
	Commissions int                    `json:"commission"`
	Total       int                    `json:"total"`
	Desc        string                 `json:"desc"`
	Info        map[string]interface{} `json:"info"`
	IsDirect    bool                   `json:"isDirect"`
	CreatedAt   time.Time              `json:"createdAt"`
}

type Ledger interface {
	WriteRecords(entries []*LedgerEntry) error
	ReadRecordsByMemberId(memberId string, cond *Condition) ([]*LedgerEntry, error)
}
