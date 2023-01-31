package main

type Member struct {
	Id           string           `json:"id"`
	Rules        map[string]*Rule `json:"rules"`
	RelationPath []string         `json:"relationPath"`
	Upstream     string           `json:"upstream"`
	Downstreams  []string         `json:"downstreams"`
}

func (m *Member) GetRule(name string) *Rule {
	if r, ok := m.Rules[name]; ok {
		return r
	}

	// Default
	return &DefaultRule
}
