package main

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

	// Initial rewards
	amount := t.Amount
	fee := t.Fee

	// Find out the member
	m, err := b.rm.GetMember(t.MemberId)
	if err != nil {
		return nil, err
	}

	// Getting default rule
	r := m.GetRule(t.Rule)

	// Create a new ledger entry for Calculating rewards for ticket owner
	le := &LedgerEntry{
		ID:          t.ID,
		Channel:     t.Channel,
		MemberId:    t.MemberId,
		Amount:      amount,
		Commissions: int64(math.Floor(float64(fee) * r.Commission)),
		Desc:        t.Desc,
		Info:        t.Info,
		IsPrimary:   true,
		PrimaryID:   t.ID,
		CreatedAt:   t.CreatedAt,
	}

	le.Total = le.Amount + le.Commissions

	entries := make([]*LedgerEntry, 0)
	entries = append(entries, le)

	fee -= le.Commissions

	// Calculating sharing and commissions by levels
	levels, err := b.GetLevels(t.MemberId)
	if err != nil {
		return nil, err
	}

	// Initializing sharing
	cormissionShare := r.Commission

	prevRule := r
	for i, l := range levels {

		// Getting default rule
		r := l.GetRule(t.Rule)

		// Create a new ledger entry for calculating feedback for upstreams
		le := &LedgerEntry{
			ID:        uuid.New().String(),
			Channel:   t.Channel,
			MemberId:  l.Id,
			Desc:      t.Desc,
			Info:      t.Info,
			IsPrimary: false,
			PrimaryID: le.PrimaryID,
			CreatedAt: t.CreatedAt,
		}

		if i != len(levels)-1 {

			//Note: Avoid precision problem
			cormissionShare = ((r.Commission * 100) - (cormissionShare * 100)) * 0.01

			// Calculating amount
			rawAmount := float64(t.Amount) * prevRule.Share
			if t.Amount < 0 {
				le.Amount = int64(math.Floor(rawAmount))
			} else {
				le.Amount = int64(math.Ceil(rawAmount))
			}

			// Calculating cormissions
			le.Commissions = int64(math.Floor(float64(t.Fee) * cormissionShare))
			le.Total = le.Amount + le.Commissions

			amount -= le.Amount
			fee -= le.Commissions

		} else {
			// The top-level agent takes the rest of amount and cormissions
			le.Amount = amount
			le.Commissions = fee
			le.Total = le.Amount + le.Commissions
		}

		entries = append(entries, le)

		prevRule = r
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
