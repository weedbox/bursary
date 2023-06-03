package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testIdCounter = 0

func genTestId() string {
	testIdCounter++
	return fmt.Sprintf("Test_%d", testIdCounter)
}

func Test_AddMembers(t *testing.T) {

	bu := NewBursary()
	defer bu.Close()

	// Add root member
	rootId := genTestId()
	err := bu.RelationManager().AddMembers([]*MemberEntry{
		&MemberEntry{
			Id: rootId,
			Rules: map[string]*Rule{
				"default": &Rule{
					Commission: 1.0,
					Share:      0.2,
				},
			},
		},
	}, "")
	assert.Nil(t, err)

	m, err := bu.RelationManager().GetMember(rootId)
	assert.Nil(t, err)
	assert.Equal(t, rootId, m.Id)

	// second level
	secondId := genTestId()
	err = bu.RelationManager().AddMembers([]*MemberEntry{
		&MemberEntry{
			Id: secondId,
			Rules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.7,
					Share:      0.3,
				},
			},
		},
	}, rootId)
	assert.Nil(t, err)

	m, err = bu.RelationManager().GetMember(secondId)
	assert.Nil(t, err)
	assert.Equal(t, secondId, m.Id)

	// third level
	thirdId := genTestId()
	err = bu.RelationManager().AddMembers([]*MemberEntry{
		&MemberEntry{
			Id: thirdId,
			Rules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.5,
					Share:      0.5,
				},
			},
		},
	}, secondId)
	assert.Nil(t, err)

	m, err = bu.RelationManager().GetMember(thirdId)
	assert.Nil(t, err)
	assert.Equal(t, thirdId, m.Id)
}

func Test_CalculateRewards(t *testing.T) {

	bu := NewBursary()
	defer bu.Close()

	levels := []*MemberEntry{
		&MemberEntry{
			Id: genTestId(),
			Rules: map[string]*Rule{
				"default": &Rule{
					Commission: 1.0,
					Share:      0,
				},
			},
		},
		&MemberEntry{
			Id: genTestId(),
			Rules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.7,
					Share:      0.7,
				},
			},
		},
		&MemberEntry{
			Id: genTestId(),
			Rules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.5,
					Share:      0.3,
				},
			},
		},
	}

	prevLevel := ""
	for _, l := range levels {
		err := bu.RelationManager().AddMembers([]*MemberEntry{
			l,
		}, prevLevel)
		assert.Nil(t, err)

		m, err := bu.RelationManager().GetMember(l.Id)
		assert.Nil(t, err)
		assert.Equal(t, l.Id, m.Id)

		prevLevel = l.Id
	}

	// Preparing a new ticket
	ticket := NewTicket()
	ticket.Rule = "default"
	ticket.MemberId = levels[2].Id
	ticket.Amount = 1000
	ticket.Fee = 50
	ticket.Total = 1050

	// Calculate rewards
	entries, err := bu.CalculateRewards(ticket)
	assert.Nil(t, err)

	// Answer
	ans := []map[string]int64{
		// level 1
		map[string]int64{
			"commission": 15,
			"income":     700,
		},
		// level 2
		map[string]int64{
			"commission": 10,
			"income":     300,
		},
		// level 3
		map[string]int64{
			"commission": 25,
			"income":     1000,
		},
	}

	for i, entry := range entries {
		a := ans[len(ans)-i-1]
		assert.Equal(t, a["commission"], entry.Commissions)
		assert.Equal(t, a["income"], entry.Amount)
		assert.Equal(t, a["income"]+a["commission"], entry.Total)
	}
}

func Test_WriteTicket(t *testing.T) {

	bu := NewBursary()
	defer bu.Close()

	// Create ledger named Slot A for testing
	err := bu.LedgerManager().Add("Channel_A", NewLedgerMemory())
	assert.Nil(t, err)

	levels := []*MemberEntry{
		&MemberEntry{
			Id: genTestId(),
			Rules: map[string]*Rule{
				"default": &Rule{
					Commission: 1.0,
					Share:      0,
				},
			},
		},
		&MemberEntry{
			Id: genTestId(),
			Rules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.7,
					Share:      0.7,
				},
			},
		},
		&MemberEntry{
			Id: genTestId(),
			Rules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.5,
					Share:      0.3,
				},
			},
		},
	}

	prevLevel := ""
	for _, l := range levels {
		err := bu.RelationManager().AddMembers([]*MemberEntry{
			l,
		}, prevLevel)
		assert.Nil(t, err)

		m, err := bu.RelationManager().GetMember(l.Id)
		assert.Nil(t, err)
		assert.Equal(t, l.Id, m.Id)

		prevLevel = l.Id
	}

	// Preparing a new ticket
	ticket := NewTicket()
	ticket.Rule = "default"
	ticket.MemberId = levels[2].Id
	ticket.Amount = 1000
	ticket.Fee = 50
	ticket.Total = 1050

	// Calculate rewards
	err = bu.WriteTicket(ticket)
	assert.Nil(t, err)

	// Answer
	ans := []map[string]int64{
		// level 1
		map[string]int64{
			"commission": 15,
			"income":     700,
		},
		// level 2
		map[string]int64{
			"commission": 10,
			"income":     300,
		},
		// level 3
		map[string]int64{
			"commission": 25,
			"income":     1000,
		},
	}

	// General ledger
	for i, l := range levels {
		records, err := bu.GeneralLedger().ReadRecordsByMemberId(l.Id, &Condition{
			Page:  1,
			Limit: 10,
		})
		assert.Nil(t, err)

		a := ans[i]
		assert.Equal(t, a["commission"], records[0].Commissions)
		assert.Equal(t, a["income"], records[0].Amount)
		assert.Equal(t, a["income"]+a["commission"], records[0].Total)
	}
}

func Test_WriteEntries(t *testing.T) {

	bu := NewBursary()
	defer bu.Close()

	// Create ledger named Slot A for testing
	err := bu.LedgerManager().Add("Channel_A", NewLedgerMemory())
	assert.Nil(t, err)

	levels := []*MemberEntry{
		&MemberEntry{
			Id: genTestId(),
			Rules: map[string]*Rule{
				"default": &Rule{
					Commission: 1.0,
					Share:      0,
				},
			},
		},
		&MemberEntry{
			Id: genTestId(),
			Rules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.7,
					Share:      0.7,
				},
			},
		},
		&MemberEntry{
			Id: genTestId(),
			Rules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.5,
					Share:      0.3,
				},
			},
		},
	}

	prevLevel := ""
	for _, l := range levels {
		err := bu.RelationManager().AddMembers([]*MemberEntry{
			l,
		}, prevLevel)
		assert.Nil(t, err)

		m, err := bu.RelationManager().GetMember(l.Id)
		assert.Nil(t, err)
		assert.Equal(t, l.Id, m.Id)

		prevLevel = l.Id
	}

	// Preparing a new ticket
	ticket := NewTicket()
	ticket.Rule = "default"
	ticket.MemberId = levels[2].Id
	ticket.Amount = 1000
	ticket.Fee = 50
	ticket.Total = 1050

	// Calculate rewards
	entries, err := bu.CalculateRewards(ticket)
	assert.Nil(t, err)

	// expected answer
	ans := []map[string]int64{
		// level 1
		map[string]int64{
			"commission": 15,
			"income":     700,
		},
		// level 2
		map[string]int64{
			"commission": 10,
			"income":     300,
		},
		// level 3
		map[string]int64{
			"commission": 25,
			"income":     1000,
		},
	}

	// Write to channel ledger
	err = bu.WriteEntries("Channel_A", entries)
	assert.Nil(t, err)

	// Check results
	for i, l := range levels {

		sl, err := bu.LedgerManager().Get("Channel_A")
		if !assert.Nil(t, err) {
			continue
		}

		records, err := sl.ReadRecordsByMemberId(l.Id, &Condition{
			Page:  1,
			Limit: 10,
		})
		if !assert.Nil(t, err) {
			continue
		}

		// Check rewards
		a := ans[i]
		assert.Equal(t, a["commission"], records[0].Commissions)
		assert.Equal(t, a["income"], records[0].Amount)
		assert.Equal(t, a["income"]+a["commission"], records[0].Total)

		// Check fields
		assert.Equal(t, ticket.ID, records[0].PrimaryID)
	}
}
