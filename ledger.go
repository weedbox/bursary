package bursary

import "time"

type LedgerEntry struct {
	Id              string                 `json:"id"`
	Channel         string                 `json:"channel"`
	MemberId        string                 `json:"member_id"`
	Contributor     string                 `json:"contributor"`
	Expense         int64                  `json:"expense"`
	Income          int64                  `json:"income"`
	Amount          int64                  `json:"amount"` // income - expense
	Fee             int64                  `json:"fee"`    // original fee
	Share           float64                `json:"share"`
	ReturnedShare   float64                `json:"returned_share"`
	CommissionShare float64                `json:"commission_share"`
	Gain            int64                  `json:"gain"`        // amount * (share + returned share by downstream)
	Commissions     int64                  `json:"commissions"` // fee * commission share
	Contributions   int64                  `json:"contributions"`
	Total           int64                  `json:"total"` // profit + commissions
	Desc            string                 `json:"desc"`
	Info            map[string]interface{} `json:"info"`
	PrimaryId       string                 `json:"primary_id"`
	IsPrimary       bool                   `json:"is_primary"`
	CreatedAt       time.Time              `json:"createdAt"`
}

type Ledger interface {
	WriteRecords(entries []*LedgerEntry) error
	ReadRecordsByMemberId(memberId string, cond *Condition) ([]*LedgerEntry, error)
}
