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

type ledger struct {
	records []*LedgerEntry
}

func NewLedger() Ledger {
	return &ledger{
		records: make([]*LedgerEntry, 0),
	}
}

func (l *ledger) WriteRecords(entries []*LedgerEntry) error {
	l.records = append(l.records, entries...)
	return nil
}

func (l *ledger) ReadRecordsByMemberId(memberId string, cond *Condition) ([]*LedgerEntry, error) {

	if cond.Page < 1 {
		cond.Page = 1
	}

	if cond.Limit < 1 {
		cond.Limit = 1
	}

	start := (cond.Page - 1) * cond.Limit

	records := make([]*LedgerEntry, 0)

	count := 0
	for i, t := range l.records {

		if i < start {
			continue
		}

		if count+1 > cond.Limit {
			break
		}

		if t.MemberId != memberId {
			continue
		}

		if cond.TimeRange != nil {
			if !t.CreatedAt.Equal(cond.TimeRange.StartTime) || !t.CreatedAt.Equal(cond.TimeRange.EndTime) ||
				!(t.CreatedAt.After(cond.TimeRange.StartTime) && t.CreatedAt.Before(cond.TimeRange.EndTime)) {
				continue
			}
		}

		records = append(records, t)
		count++
	}

	return records, nil
}
