package bursary

import (
	"math"

	"github.com/google/uuid"
)

type Bursary interface {
	RelationManager() RelationManager
	LedgerManager() LedgerManager
	GeneralLedger() Ledger
	GetLevels(memberId string) ([]*Member, error)
	CalculateRewards(t *Ticket) ([]*LedgerEntry, error)
	WriteTicket(t *Ticket) error
	WriteEntry(le *LedgerEntry) error
	WriteEntries(ledgerName string, entries []*LedgerEntry) error
	Close() error
}

type bursary struct {
	rm RelationManager
	lm LedgerManager
	gl Ledger
}

type Opt func(*bursary)

func NewBursary(opts ...Opt) Bursary {

	b := &bursary{}

	for _, opt := range opts {
		opt(b)
	}

	if b.rm == nil {
		// Using memory to store relationship by default
		b.rm = NewRelationManagerMemory()
	}

	if b.lm == nil {
		b.lm = NewLedgerManager()
	}

	if b.gl == nil {
		// Using memory to store ledger by default
		b.gl = NewLedgerMemory()
		b.lm.Add("general", b.gl)
	}

	return b
}

func WithRelationManager(rm RelationManager) Opt {
	return func(b *bursary) {
		b.rm = rm
	}
}

func WithLedgerManager(lm LedgerManager) Opt {
	return func(b *bursary) {
		b.lm = lm
	}
}

func WithGeneralLedger(l Ledger) Opt {
	return func(b *bursary) {
		b.gl = l
	}
}

func (b *bursary) RelationManager() RelationManager {
	return b.rm
}

func (b *bursary) LedgerManager() LedgerManager {
	return b.lm
}

func (b *bursary) GeneralLedger() Ledger {
	return b.gl
}

func (b *bursary) Close() error {
	b.rm.Close()
	return nil
}

func (b *bursary) AddMember(memberEntry *MemberEntry, upstream string) error {
	return b.rm.AddMembers([]*MemberEntry{
		memberEntry,
	}, upstream)
}

func (b *bursary) GetLevels(memberId string) ([]*Member, error) {

	upstreams, err := b.rm.GetUpstreams(memberId)
	if err != nil {
		return nil, err
	}

	// Reverse upstreams list
	levels := make([]*Member, 0)
	for i := len(upstreams) - 1; i >= 0; i-- {
		levels = append(levels, upstreams[i])
	}

	return levels, nil
}

func (b *bursary) CalculateRewards(t *Ticket) ([]*LedgerEntry, error) {

	// Find out the edge member
	m, err := b.rm.GetMember(t.MemberID)
	if err != nil {
		return nil, err
	}

	// Getting rule for specific channel
	r := m.GetChannelRule(t.Channel)
	if r == nil {
		// Using default rule if it doesn't exist
		r = &DefaultRule
	}

	// Create a new ledger entry for Calculating rewards for ticket owner
	le := &LedgerEntry{
		ID:              t.ID,
		Channel:         t.Channel,
		MemberID:        t.MemberID,
		Contributor:     t.MemberID, // self
		Expense:         t.Expense,
		Income:          t.Income,
		Fee:             t.Fee,
		Amount:          t.Amount,
		Share:           r.Share,
		ReturnedShare:   0.0,
		CommissionShare: r.Commission,
		Desc:            t.Desc,
		Info:            t.Info,
		IsPrimary:       true,
		PrimaryID:       t.ID,
		CreatedAt:       t.CreatedAt,
	}

	// Calculate gain and commissions
	le.Commissions = int64(math.Floor(float64(t.Fee) * r.Commission))
	le.Gain = int64(math.Floor(float64(t.Amount) * r.Share))
	le.Contributions = t.Amount - le.Gain

	// Deduct the delivered parts
	fee := t.Fee - le.Commissions

	le.Total = le.Gain + le.Commissions

	// Add entry of ticket owner to list
	entries := make([]*LedgerEntry, 0)
	entries = append(entries, le)

	// Getting all levels from edge to root
	levels, err := b.GetLevels(t.MemberID)
	if err != nil {
		return nil, err
	}

	// Calculating sharing and commissions by levels
	downstreamEntry := le
	downstreamRule := r
	for i, l := range levels {

		downstreamEntry.Upstream = l.ID

		// Getting default rule
		r := l.GetChannelRule(t.Channel)
		if r == nil {
			// Using pervious rule if it doesn't exist
			r = &Rule{
				Commission: downstreamEntry.CommissionShare,
				Share:      0.0,
			}
		}

		// Create a new ledger entry for calculating feedback for upstreams
		le := &LedgerEntry{
			ID:              uuid.New().String(),
			Channel:         t.Channel,
			MemberID:        l.ID,
			Contributor:     downstreamEntry.ID,
			Expense:         t.Expense,
			Income:          t.Income,
			Amount:          t.Amount,
			Share:           r.Share,
			ReturnedShare:   0.0,
			CommissionShare: r.Commission,
			Desc:            t.Desc,
			Info:            t.Info,
			IsPrimary:       false,
			PrimaryID:       le.PrimaryID,
			CreatedAt:       t.CreatedAt,
		}

		if i != len(levels)-1 {

			// Calculate gain and commissions shares
			commissionShare := (r.Commission*100 - downstreamEntry.CommissionShare*100) / 100
			share := (r.Share*100 + downstreamEntry.ReturnedShare*100 - downstreamEntry.Share*100 - downstreamRule.ReturnedShare*100) / 100

			// Return share to upstream
			le.ReturnedShare = downstreamRule.ReturnedShare

			// Calculate gain and commissions
			le.Commissions = int64(math.Floor(float64(t.Fee) * commissionShare))
			le.Gain = int64(math.Floor(float64(t.Amount) * share))

			fee -= le.Commissions

		} else {
			// The top-level agent takes the rest of contributions and cormissions
			le.Gain = downstreamEntry.Contributions
			le.Commissions = fee
		}

		le.Contributions = downstreamEntry.Contributions - le.Gain
		le.Total = le.Gain + le.Commissions

		entries = append(entries, le)

		downstreamEntry = le
		downstreamRule = r
	}

	return entries, nil
}

func (b *bursary) WriteTicket(t *Ticket) error {

	entries, err := b.CalculateRewards(t)
	if err != nil {
		return err
	}

	// Write reward results to general ledger
	return b.gl.WriteRecords(entries)
}

func (b *bursary) WriteEntry(le *LedgerEntry) error {

	// Attempt to find ledger for specific channel
	l, err := b.lm.Get(le.Channel)
	if err != nil {
		return err
	}

	return l.WriteRecords([]*LedgerEntry{le})
}

func (b *bursary) WriteEntries(ledgerName string, entries []*LedgerEntry) error {

	// Attempt to find ledger for specific channel
	l, err := b.lm.Get(ledgerName)
	if err != nil {
		return err
	}

	return l.WriteRecords(entries)
}
