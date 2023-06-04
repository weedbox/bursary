package bursary

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_RelationManager_ListMembers(t *testing.T) {

	bu := NewBursary()
	defer bu.Close()

	levels := []*MemberEntry{
		&MemberEntry{
			Id: genTestId(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 1.0,
					Share:      0,
				},
			},
		},
		&MemberEntry{
			Id: genTestId(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.7,
					Share:      0.7,
				},
			},
		},
		&MemberEntry{
			Id: genTestId(),
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

		m, err := bu.RelationManager().GetMember(l.Id)
		assert.Nil(t, err)
		assert.Equal(t, l.Id, m.Id)

		prevLevel = l.Id
	}

	members, err := bu.RelationManager().ListMembers("", &Condition{
		Page:  1,
		Limit: 2,
	})
	assert.Nil(t, err)
	assert.Len(t, members, 1)

	members, err = bu.RelationManager().ListMembers("", &Condition{
		Page:  2,
		Limit: 1,
	})
	assert.Nil(t, err)
	assert.Len(t, members, 0)
}

func Test_RelationManager_RemoveChannel(t *testing.T) {

	bu := NewBursary()
	defer bu.Close()

	levels := []*MemberEntry{
		&MemberEntry{
			Id: genTestId(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 1.0,
					Share:      0,
				},
			},
		},
		&MemberEntry{
			Id: genTestId(),
			ChannelRules: map[string]*Rule{
				"default": &Rule{
					Commission: 0.7,
					Share:      0.7,
				},
			},
		},
		&MemberEntry{
			Id: genTestId(),
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

		m, err := bu.RelationManager().GetMember(l.Id)
		assert.Nil(t, err)
		assert.Equal(t, l.Id, m.Id)

		prevLevel = l.Id
	}

	err := bu.RelationManager().RemoveChannel("default")
	assert.Nil(t, err)

	members, err := bu.RelationManager().ListMembers("", &Condition{
		Page:  1,
		Limit: 10,
	})
	assert.Nil(t, err)

	for _, m := range members {
		r := m.GetChannelRule("default")
		assert.Nil(t, r)

	}
}
