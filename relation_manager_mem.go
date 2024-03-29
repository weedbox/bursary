package bursary

type relationManagerMemory struct {
	members map[string]*Member
}

func NewRelationManagerMemory() RelationManager {
	return &relationManagerMemory{
		members: make(map[string]*Member),
	}
}

func (rm *relationManagerMemory) Close() error {
	return nil
}

func (rm *relationManagerMemory) GetPath(mid string) ([]string, error) {

	p := make([]string, 0)
	if len(mid) != 0 {

		// find upstream to get relation path
		usm, err := rm.GetMember(mid)
		if err != nil {
			return p, ErrUpstreamNotFound
		}

		p = append(p, usm.RelationPath...)
		p = append(p, mid)
	}

	return p, nil
}

func (rm *relationManagerMemory) ChangePath(mid string, newPath []string) error {

	m, err := rm.GetMember(mid)
	if err != nil {
		return ErrMemberNotFound
	}

	m.RelationPath = newPath

	return nil
}

func (rm *relationManagerMemory) GetMember(mid string) (*Member, error) {

	if m, ok := rm.members[mid]; ok {
		return m, nil
	}

	return nil, ErrMemberNotFound
}

func (rm *relationManagerMemory) AddMembers(members []*MemberEntry, upstream string) error {

	rp, err := rm.GetPath(upstream)
	if err != nil {
		return ErrUpstreamNotFound
	}

	for _, me := range members {

		m := &Member{
			ID:           me.ID,
			ChannelRules: me.ChannelRules,
			RelationPath: rp,
			Upstream:     upstream,
		}

		rm.members[m.ID] = m
	}

	return nil
}

func (rm *relationManagerMemory) MoveMembers(mids []string, upstream string) error {

	// Getting all members
	for _, mid := range mids {

		m, err := rm.GetMember(mid)
		if err != nil {
			return err
		}

		rp, err := rm.GetPath(upstream)
		if err != nil {
			return ErrUpstreamNotFound
		}

		m.Upstream = upstream
		m.RelationPath = rp

		curPath := append(rp, mid)

		// find and update downstreams
		for _, ds := range rm.members {
			if ds.Upstream == m.ID {
				rm.ChangePath(ds.ID, curPath)
			}
		}
	}

	return nil
}

func (rm *relationManagerMemory) DeleteMembers(mids []string) error {

	for _, mid := range mids {
		delete(rm.members, mid)
	}

	return nil
}

func (rm *relationManagerMemory) GetUpstreams(mid string) ([]*Member, error) {

	members := make([]*Member, 0)

	m, err := rm.GetMember(mid)
	if err != nil {
		return members, err
	}

	// Getting all members according to relation path
	for _, usID := range m.RelationPath {

		usm, err := rm.GetMember(usID)
		if err != nil {
			return nil, err
		}

		members = append(members, usm)
	}

	return members, nil
}

func (rm *relationManagerMemory) ListMembers(upstream string, cond *Condition) ([]*Member, error) {

	if cond.Page < 1 {
		cond.Page = 1
	}

	if cond.Limit < 1 {
		cond.Limit = 1
	}

	start := (cond.Page - 1) * cond.Limit

	members := make([]*Member, 0)

	count := 0
	cur := 0
	for _, m := range rm.members {

		if m.Upstream != upstream {
			continue
		}

		if cur < start {
			cur++
			continue
		}

		if count+1 > cond.Limit {
			break
		}

		members = append(members, m)

		count++
	}

	return members, nil
}

func (rm *relationManagerMemory) UpdateChannelRule(mid string, channel string, rule *Rule) error {

	m, err := rm.GetMember(mid)
	if err != nil {
		return err
	}

	m.ChannelRules[channel] = rule

	return nil
}

func (rm *relationManagerMemory) RemoveChannelRule(mid string, channel string) error {

	m, err := rm.GetMember(mid)
	if err != nil {
		return err
	}

	delete(m.ChannelRules, channel)

	return nil
}

func (rm *relationManagerMemory) RemoveChannel(channel string) error {

	// Remove specific channel rule from all members
	for _, m := range rm.members {
		delete(m.ChannelRules, channel)
	}

	return nil
}
