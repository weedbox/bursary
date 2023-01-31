package main

import (
	"errors"
)

var (
	ErrMemberNotFound   = errors.New("bursary: member not found")
	ErrUpstreamNotFound = errors.New("bursary: upstream not found")
)

type MemberEntry struct {
	Id    string           `json:"id"`
	Rules map[string]*Rule `json:"rules"`
}

type RelationManager interface {
	AddMembers(members []*MemberEntry, upstream string) error
	GetMember(mid string) (*Member, error)
	MoveMembers(mids []string, upstream string) error
	DeleteMembers(mids []string) error
	GetUpstreams(mid string) ([]*Member, error)
	ListMembers(cond *Condition) ([]*Member, error)
	UpdateRule(mid string, ruleName string, rule *Rule) error
	RemoveRule(mid string, ruleName string) error
	Close() error
}

type relationManager struct {
	members map[string]*Member
}

func NewRelationManager() RelationManager {
	return &relationManager{
		members: make(map[string]*Member),
	}
}

func (rm *relationManager) Close() error {
	return nil
}

func (rm *relationManager) GetPath(mid string) ([]string, error) {

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

func (rm *relationManager) ChangePath(mid string, newPath []string) error {

	m, err := rm.GetMember(mid)
	if err != nil {
		return ErrMemberNotFound
	}

	m.RelationPath = newPath

	return nil
}

func (rm *relationManager) GetMember(mid string) (*Member, error) {

	if m, ok := rm.members[mid]; ok {
		return m, nil
	}

	return nil, ErrMemberNotFound
}

func (rm *relationManager) AddMembers(members []*MemberEntry, upstream string) error {

	rp, err := rm.GetPath(upstream)
	if err != nil {
		return ErrUpstreamNotFound
	}

	for _, me := range members {

		m := &Member{
			Id:           me.Id,
			Rules:        me.Rules,
			RelationPath: rp,
			Upstream:     upstream,
			Downstreams:  make([]string, 0),
		}

		rm.members[m.Id] = m
	}

	return nil
}

func (rm *relationManager) MoveMembers(mids []string, upstream string) error {

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

		// update downstreams
		for _, dsid := range m.Downstreams {
			rm.ChangePath(dsid, curPath)
		}
	}

	return nil
}

func (rm *relationManager) DeleteMembers(mids []string) error {

	for _, mid := range mids {
		delete(rm.members, mid)
	}

	return nil
}

func (rm *relationManager) GetUpstreams(mid string) ([]*Member, error) {

	members := make([]*Member, 0)

	m, err := rm.GetMember(mid)
	if err != nil {
		return members, err
	}

	// Getting all members according to relation path
	for _, usId := range m.RelationPath {

		usm, err := rm.GetMember(usId)
		if err != nil {
			return nil, err
		}

		members = append(members, usm)
	}

	return members, nil
}

func (rm *relationManager) ListMembers(cond *Condition) ([]*Member, error) {

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

func (rm *relationManager) UpdateRule(mid string, ruleName string, rule *Rule) error {

	m, err := rm.GetMember(mid)
	if err != nil {
		return err
	}

	m.Rules[ruleName] = rule

	return nil
}

func (rm *relationManager) RemoveRule(mid string, ruleName string) error {

	m, err := rm.GetMember(mid)
	if err != nil {
		return err
	}

	delete(m.Rules, ruleName)

	return nil
}
