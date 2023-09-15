package bursary

type Member struct {
	ID           string           `json:"id"`
	ChannelRules map[string]*Rule `json:"channel_rules"`
	RelationPath []string         `json:"relation_path"`
	Upstream     string           `json:"upstream"`
}

func (m *Member) GetChannelRule(channel string) *Rule {

	// Finding the rule for specific channel
	if r, ok := m.ChannelRules[channel]; ok {
		return r
	}

	return nil
}
