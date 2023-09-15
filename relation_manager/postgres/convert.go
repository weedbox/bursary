package relation_manager_postgres

import "github.com/weedbox/bursary"

func (mr *MemberRecord) ToMemberObject() *bursary.Member {

	m := &bursary.Member{
		ID:           mr.ID,
		ChannelRules: make(map[string]*bursary.Rule),
		RelationPath: mr.RelationPath,
		Upstream:     mr.Upstream,
	}

	for channel, rule := range mr.ChannelRules {
		m.ChannelRules[channel] = &bursary.Rule{
			Commission: rule.Commission,
			Share:      rule.Share,
		}
	}

	return m
}
