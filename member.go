package main

type Member struct {
	Id           string           `json:"id"`
	ChannelRules map[string]*Rule `json:"channelRules"`
	RelationPath []string         `json:"relationPath"`
	Upstream     string           `json:"upstream"`
}

func (m *Member) GetChannelRule(channel string) *Rule {

	// Finding the rule for specific channel
	if r, ok := m.ChannelRules[channel]; ok {
		return r
	}

	return nil
}
