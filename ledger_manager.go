package bursary

import "errors"

var (
	ErrLedgerNotFound = errors.New("bursary: ledger not found")
)

type LedgerManager interface {
	Add(name string, l Ledger) error
	Get(name string) (Ledger, error)
	Delete(name string) error
}

type ledgerManager struct {
	ledgers map[string]Ledger
}

func NewLedgerManager() LedgerManager {
	return &ledgerManager{
		ledgers: make(map[string]Ledger),
	}
}

func (lm *ledgerManager) Add(name string, l Ledger) error {
	lm.ledgers[name] = l
	return nil
}

func (lm *ledgerManager) Get(name string) (Ledger, error) {

	if l, ok := lm.ledgers[name]; ok {
		return l, nil
	}

	return nil, ErrLedgerNotFound
}

func (lm *ledgerManager) Delete(name string) error {
	delete(lm.ledgers, name)
	return nil
}
