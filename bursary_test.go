package bursary

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testIDCounter = 0

func genTestID() string {
	testIDCounter++
	return fmt.Sprintf("Test_%d", testIDCounter)
}

func Test_AddMembers(t *testing.T) {

	bu := NewBursary()
	defer bu.Close()

	// Add root member
	rootID := genTestID()
	err := bu.RelationManager().AddMembers([]*MemberEntry{
		&MemberEntry{
			ID: rootID,
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 1.0,
					Share:      1.0,
				},
			},
		},
	}, "")
	assert.Nil(t, err)

	m, err := bu.RelationManager().GetMember(rootID)
	assert.Nil(t, err)
	assert.Equal(t, rootID, m.ID)

	// second level
	secondID := genTestID()
	err = bu.RelationManager().AddMembers([]*MemberEntry{
		&MemberEntry{
			ID: secondID,
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.7,
					Share:      0.8,
				},
			},
		},
	}, rootID)
	assert.Nil(t, err)

	m, err = bu.RelationManager().GetMember(secondID)
	assert.Nil(t, err)
	assert.Equal(t, secondID, m.ID)

	// third level
	thirdID := genTestID()
	err = bu.RelationManager().AddMembers([]*MemberEntry{
		&MemberEntry{
			ID: thirdID,
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.5,
					Share:      0.5,
				},
			},
		},
	}, secondID)
	assert.Nil(t, err)

	m, err = bu.RelationManager().GetMember(thirdID)
	assert.Nil(t, err)
	assert.Equal(t, thirdID, m.ID)
}

func Test_CalculateRewards(t *testing.T) {

	bu := NewBursary()
	defer bu.Close()

	levels := []*MemberEntry{
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 1.0,
					Share:      1.0,
				},
			},
		},
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.7,
					Share:      0.8,
				},
			},
		},
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
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

		m, err := bu.RelationManager().GetMember(l.ID)
		assert.Nil(t, err)
		assert.Equal(t, l.ID, m.ID)

		prevLevel = l.ID
	}

	// Preparing a new ticket
	ticket := NewTicket()
	ticket.Channel = "default"
	ticket.MemberID = levels[2].ID
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
			"commission":    15,
			"gain":          200,
			"contributions": 0, // 200 - 200
		},
		// level 2
		map[string]int64{
			"commission":    10,
			"gain":          500,
			"contributions": 200, // 700 - 500
		},
		// level 3
		map[string]int64{
			"commission":    25,
			"gain":          300,
			"contributions": 700, // 1000 - 300
		},
	}

	for i, entry := range entries {
		a := ans[len(ans)-i-1]
		assert.Equal(t, a["commission"], entry.Commissions)
		assert.Equal(t, a["gain"], entry.Gain)
		assert.Equal(t, a["contributions"], entry.Contributions)

		if entry.IsPrimary {
			assert.Equal(t, ticket.Amount-a["gain"]+a["commission"], entry.Total)
		} else {
			assert.Equal(t, a["gain"]+a["commission"], entry.Total)
		}
	}
}

func Test_CalculateRewards_Aliquant(t *testing.T) {

	bu := NewBursary()
	defer bu.Close()

	levels := []*MemberEntry{
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 1.0,
					Share:      1.0,
				},
			},
		},
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.7,
					Share:      0.6,
				},
			},
		},
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
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

		m, err := bu.RelationManager().GetMember(l.ID)
		assert.Nil(t, err)
		assert.Equal(t, l.ID, m.ID)

		prevLevel = l.ID
	}

	// Preparing a new ticket
	ticket := NewTicket()
	ticket.Channel = "default"
	ticket.MemberID = levels[2].ID
	ticket.Amount = 999
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
			"gain":       401,
		},
		// level 2
		map[string]int64{
			"commission": 10,
			"gain":       299, // 299.7
		},
		// level 3
		map[string]int64{
			"commission": 25,
			"gain":       299, // 299.7
		},
	}

	for i, entry := range entries {
		a := ans[len(ans)-i-1]
		assert.Equal(t, a["commission"], entry.Commissions)
		assert.Equal(t, a["gain"], entry.Gain)

		if entry.IsPrimary {
			assert.Equal(t, ticket.Amount-a["gain"]+a["commission"], entry.Total)
		} else {
			assert.Equal(t, a["gain"]+a["commission"], entry.Total)
		}
	}
}

func Test_CalculateRewards_Negative(t *testing.T) {

	bu := NewBursary()
	defer bu.Close()

	levels := []*MemberEntry{
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 1.0,
					Share:      1.0,
				},
			},
		},
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.7,
					Share:      0.6,
				},
			},
		},
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
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

		m, err := bu.RelationManager().GetMember(l.ID)
		assert.Nil(t, err)
		assert.Equal(t, l.ID, m.ID)

		prevLevel = l.ID
	}

	// Preparing a new ticket
	ticket := NewTicket()
	ticket.Channel = "default"
	ticket.MemberID = levels[2].ID
	ticket.Amount = -999
	ticket.Fee = 50

	// Calculate rewards
	entries, err := bu.CalculateRewards(ticket)
	assert.Nil(t, err)

	// Answer
	ans := []map[string]int64{
		// level 1
		map[string]int64{
			"commission":    15,
			"gain":          -399,
			"contributions": 0, // -399 + 399
		},
		// level 2
		map[string]int64{
			"commission":    10,
			"gain":          -300, // -299.7
			"contributions": -399, // -699 + 300
		},
		// level 3
		map[string]int64{
			"commission":    25,
			"gain":          -300, // -299.7
			"contributions": -699, // -999 + 300
		},
	}

	for i, entry := range entries {
		a := ans[len(ans)-i-1]
		assert.Equal(t, a["commission"], entry.Commissions)
		assert.Equal(t, a["gain"], entry.Gain)
		assert.Equal(t, a["contributions"], entry.Contributions)

		if entry.IsPrimary {
			assert.Equal(t, ticket.Amount-a["gain"]+a["commission"], entry.Total)
		} else {
			assert.Equal(t, a["gain"]+a["commission"], entry.Total)
		}
	}
}

func Test_CalculateRewards_ReturnShare(t *testing.T) {

	bu := NewBursary()
	defer bu.Close()

	levels := []*MemberEntry{
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 1.0,
					Share:      1.0,
				},
			},
		},
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.7,
					Share:      0.9,
				},
			},
		},
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.7,
					Share:      0.8,
				},
			},
		},
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission:    0.5,
					Share:         0.3,
					ReturnedShare: 0.4, // upstream should keep 10% only
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

		m, err := bu.RelationManager().GetMember(l.ID)
		assert.Nil(t, err)
		assert.Equal(t, l.ID, m.ID)

		prevLevel = l.ID
	}

	// Preparing a new ticket
	ticket := NewTicket()
	ticket.Channel = "default"
	ticket.MemberID = levels[len(levels)-1].ID
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
			"gain":       100,
		},
		// level 2
		map[string]int64{
			"commission": 0,
			"gain":       500,
		},
		// level 3
		map[string]int64{
			"commission": 10,
			"gain":       100,
		},
		// level 4
		map[string]int64{
			"commission": 25,
			"gain":       300,
		},
	}

	for i, entry := range entries {
		a := ans[len(ans)-i-1]
		assert.Equal(t, a["commission"], entry.Commissions)
		assert.Equal(t, a["gain"], entry.Gain)

		if entry.IsPrimary {
			assert.Equal(t, ticket.Amount-a["gain"]+a["commission"], entry.Total)
		} else {
			assert.Equal(t, a["gain"]+a["commission"], entry.Total)
		}
	}
}

func Test_WriteTicket(t *testing.T) {

	bu := NewBursary()
	defer bu.Close()

	// Create ledger named default for testing channel
	err := bu.LedgerManager().Add("default", NewLedgerMemory())
	assert.Nil(t, err)

	levels := []*MemberEntry{
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 1.0,
					Share:      1.0,
				},
			},
		},
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.7,
					Share:      0.9,
				},
			},
		},
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
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

		m, err := bu.RelationManager().GetMember(l.ID)
		assert.Nil(t, err)
		assert.Equal(t, l.ID, m.ID)

		prevLevel = l.ID
	}

	// Preparing a new ticket
	ticket := NewTicket()
	ticket.Channel = "default"
	ticket.MemberID = levels[2].ID
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
			"gain":       100,
		},
		// level 2
		map[string]int64{
			"commission": 10,
			"gain":       600,
		},
		// level 3
		map[string]int64{
			"commission": 25,
			"gain":       300,
		},
	}

	// General ledger
	for i, l := range levels {
		records, err := bu.GeneralLedger().ReadRecordsByMemberID(l.ID, &Condition{
			Page:  1,
			Limit: 10,
		})
		assert.Nil(t, err)

		a := ans[i]
		assert.Equal(t, a["commission"], records[0].Commissions)
		assert.Equal(t, a["gain"], records[0].Gain)

		if records[0].IsPrimary {
			assert.Equal(t, ticket.Amount-a["gain"]+a["commission"], records[0].Total)
		} else {
			assert.Equal(t, a["gain"]+a["commission"], records[0].Total)
		}
	}
}

func Test_WriteEntries(t *testing.T) {

	bu := NewBursary()
	defer bu.Close()

	// Create ledger named default for testing channel
	err := bu.LedgerManager().Add("default", NewLedgerMemory())
	assert.Nil(t, err)

	levels := []*MemberEntry{
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 1.0,
					Share:      1.0,
				},
			},
		},
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.7,
					Share:      1.0,
				},
			},
		},
		&MemberEntry{
			ID: genTestID(),
			ChannelRules: map[string]*Rule{
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

		m, err := bu.RelationManager().GetMember(l.ID)
		assert.Nil(t, err)
		assert.Equal(t, l.ID, m.ID)

		prevLevel = l.ID
	}

	// Preparing a new ticket
	ticket := NewTicket()
	ticket.Channel = "default"
	ticket.MemberID = levels[2].ID
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
			"gain":       0,
		},
		// level 2
		map[string]int64{
			"commission": 10,
			"gain":       700,
		},
		// level 3
		map[string]int64{
			"commission": 25,
			"gain":       300,
		},
	}

	// Write to channel ledger
	err = bu.WriteEntries("default", entries)
	assert.Nil(t, err)

	// Check results
	for i, l := range levels {

		sl, err := bu.LedgerManager().Get("default")
		if !assert.Nil(t, err) {
			continue
		}

		records, err := sl.ReadRecordsByMemberID(l.ID, &Condition{
			Page:  1,
			Limit: 10,
		})
		if !assert.Nil(t, err) {
			continue
		}

		// Check rewards
		a := ans[i]
		assert.Equal(t, a["commission"], records[0].Commissions)
		assert.Equal(t, a["gain"], records[0].Gain)

		if records[0].IsPrimary {
			assert.Equal(t, ticket.Amount-a["gain"]+a["commission"], records[0].Total)
		} else {
			assert.Equal(t, a["gain"]+a["commission"], records[0].Total)
		}

		// Check fields
		assert.Equal(t, ticket.ID, records[0].PrimaryID)
	}
}
