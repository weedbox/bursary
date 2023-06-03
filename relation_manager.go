package bursary

import (
	"errors"
)

var (
	ErrMemberNotFound   = errors.New("bursary: member not found")
	ErrUpstreamNotFound = errors.New("bursary: upstream not found")
)

type MemberEntry struct {
	Id           string           `json:"id"`
	ChannelRules map[string]*Rule `json:"channelRules"`
}

type RelationManager interface {
	AddMembers(members []*MemberEntry, upstream string) error
	GetMember(mid string) (*Member, error)
	MoveMembers(mids []string, upstream string) error
	DeleteMembers(mids []string) error
	GetUpstreams(mid string) ([]*Member, error)
	ListMembers(cond *Condition) ([]*Member, error)
	UpdateChannelRule(mid string, channel string, rule *Rule) error
	RemoveChannelRule(mid string, channel string) error
	RemoveChannel(channel string) error
	Close() error
}
