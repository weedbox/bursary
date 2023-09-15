package bursary

type ledgerMemory struct {
	records []*LedgerEntry
}

func NewLedgerMemory() Ledger {
	return &ledgerMemory{
		records: make([]*LedgerEntry, 0),
	}
}

func (l *ledgerMemory) WriteRecords(entries []*LedgerEntry) error {
	l.records = append(l.records, entries...)
	return nil
}

func (l *ledgerMemory) ReadRecordsByMemberID(memberID string, cond *Condition) ([]*LedgerEntry, error) {

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

		if t.MemberID != memberID {
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
