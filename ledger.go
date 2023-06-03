package main

import "time"

type LedgerEntry struct {
	ID          string                 `json:"id"`
	Channel     string                 `json:"channel"`
	MemberId    string                 `json:"memberId"`
	Amount      int64                  `json:"amount"`
	Commissions int64                  `json:"commission"`
	Total       int64                  `json:"total"`
	Desc        string                 `json:"desc"`
	Info        map[string]interface{} `json:"info"`
	PrimaryID   string                 `json:"primaryId"`
	IsPrimary   bool                   `json:"isPrimary"`
	CreatedAt   time.Time              `json:"createdAt"`
}

type Ledger interface {
	WriteRecords(entries []*LedgerEntry) error
	ReadRecordsByMemberId(memberId string, cond *Condition) ([]*LedgerEntry, error)
}
