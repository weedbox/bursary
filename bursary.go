package main

import (
	"math"
)

type Bursary interface {
	RelationManager() RelationManager
	LedgerManager() LedgerManager
	GeneralLedger() Ledger
	GetLevels(memberId string) ([]*Member, error)
	CalculateRewards(t *Ticket) ([]*LedgerEntry, error)
	CreateTicket(t *Ticket) error
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
		b.rm = NewRelationManager()
	}

	if b.lm == nil {
		b.lm = NewLedgerManager()
	}

	if b.gl == nil {
		b.gl = NewLedger()
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

	// Initial
	amount := t.Amount
	fee := t.Fee

	m, err := b.rm.GetMember(t.MemberId)
	if err != nil {
		return nil, err
	}

	// Getting default rule
	r := m.GetRule(t.Rule)

	// Create a new ledger entry for Calculating rewards for ticket owner
	le := &LedgerEntry{
		MemberId:    t.MemberId,
		Amount:      amount,
		Commissions: int(math.Floor(float64(fee) * r.Commission)),
		Desc:        t.Desc,
		Info:        t.Info,
		IsDirect:    true,
		CreatedAt:   t.CreatedAt,
	}

	le.Total = le.Amount + le.Commissions

	entries := make([]*LedgerEntry, 0)
	entries = append(entries, le)

	fee -= le.Commissions

	// Calculating sharing and cormissions by levels
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

		le := &LedgerEntry{
			MemberId:  l.Id,
			Desc:      t.Desc,
			Info:      t.Info,
			IsDirect:  false,
			CreatedAt: t.CreatedAt,
		}

		if i != len(levels)-1 {

			//Note: Avoid precision problem
			cormissionShare = ((r.Commission * 100) - (cormissionShare * 100)) * 0.01

			// Calculating amount and cormissions by rules
			if t.Amount < 0 {
				le.Amount = int(math.Floor(float64(t.Amount) * prevRule.Share))
			} else {
				le.Amount = int(math.Ceil(float64(t.Amount) * prevRule.Share))
			}

			le.Commissions = int(math.Floor(float64(t.Fee) * cormissionShare))
			le.Total = le.Amount + le.Commissions

			amount -= le.Amount
			fee -= le.Commissions

		} else {
			// The top-level agent takes the rest of income and cormissions
			le.Amount = amount
			le.Commissions = fee
			le.Total = le.Amount + le.Commissions
		}

		entries = append(entries, le)

		prevRule = r
	}

	return entries, nil
}

func (b *bursary) CreateTicket(t *Ticket) error {

	entries, err := b.CalculateRewards(t)
	if err != nil {
		return err
	}

	// Write entries to general ledger
	err = b.gl.WriteRecords(entries)
	if err != nil {
		return err
	}

	// No need to write to other ledger
	if len(t.LedgerId) == 0 {
		return nil
	}

	// Write to specifc ledger
	l, err := b.lm.Get(t.LedgerId)
	if err != nil {
		return err
	}

	return l.WriteRecords(entries)
}
