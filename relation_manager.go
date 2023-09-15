package bursary

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrMemberRequired   = errors.New("bursary: require member")
	ErrMemberNotFound   = errors.New("bursary: member not found")
	ErrUpstreamNotFound = errors.New("bursary: upstream not found")
)

type MemberEntry struct {
	ID           string           `json:"id"`
	ChannelRules map[string]*Rule `json:"channel_rules"`
}

type RelationManager interface {
	AddMembers(members []*MemberEntry, upstream string) error
	ChangePath(mid string, newPath []string) error
	DeleteMembers(mids []string) error
	GetPath(mid string) ([]string, error)
	GetMember(mid string) (*Member, error)
	GetUpstreams(mid string) ([]*Member, error)
	MoveMembers(mids []string, upstream string) error
	ListMembers(upstream string, cond *Condition) ([]*Member, error)
	UpdateChannelRule(mid string, channel string, rule *Rule) error
	RemoveChannelRule(mid string, channel string) error
	RemoveChannel(channel string) error
	Close() error
}

func NewMemberEntry() *MemberEntry {
	return &MemberEntry{
		ID:           uuid.New().String(),
		ChannelRules: make(map[string]*Rule),
	}
}
