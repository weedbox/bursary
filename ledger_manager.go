package main

import "errors"

var (
	ErrLedgerNotFound = errors.New("bursary: ledger not found")
)

type LedgerManager interface {
	Add(ledgerId string, l Ledger) error
	Get(ledgerId string) (Ledger, error)
	Delete(ledgerId string) error
}

type ledgerManager struct {
	ledgers map[string]Ledger
}

func NewLedgerManager() LedgerManager {
	return &ledgerManager{
		ledgers: make(map[string]Ledger),
	}
}

func (lm *ledgerManager) Add(ledgerId string, l Ledger) error {
	lm.ledgers[ledgerId] = l
	return nil
}

func (lm *ledgerManager) Get(ledgerId string) (Ledger, error) {

	if l, ok := lm.ledgers[ledgerId]; ok {
		return l, nil
	}

	return nil, ErrLedgerNotFound
}

func (lm *ledgerManager) Delete(ledgerId string) error {
	delete(lm.ledgers, ledgerId)
	return nil
}
