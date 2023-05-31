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
