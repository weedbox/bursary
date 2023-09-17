package relation_manager_postgres

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weedbox/bursary"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var testDb *sqlx.DB
var testTable = "relationships_test"
var testRM *RelationManagerPostgres
var testBu bursary.Bursary

func init() {

	// Connect to postgres server
	db, err := sqlx.Connect("postgres", "port=32768 user=postgres password=1qazXSW@ dbname=bursary sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}

	testDb = db

	rm := NewRelationManagerPostgres(
		WithDb(testDb),
		WithTableName(testTable),
	)

	err = rm.Init()
	if err != nil {
		log.Fatalln(err)
	}

	testRM = rm

	// Initialize bursary
	testBu = bursary.NewBursary(
		bursary.WithRelationManager(testRM),
	)
}

func uninit() {
	cmd := fmt.Sprintf(`TRUNCATE TABLE %s`, testTable)
	_, err := testDb.Exec(cmd)
	if err != nil {
		log.Fatalln(err)
	}
}

func Test_RelationManagerPostgres_AddMembers(t *testing.T) {

	defer uninit()

	var levels []*bursary.MemberEntry

	// Preparing members
	me := bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 1.0,
		Share:      0,
	}
	levels = append(levels, me)

	me = bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 0.7,
		Share:      0.7,
	}
	levels = append(levels, me)

	me = bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 0.5,
		Share:      0.3,
	}
	levels = append(levels, me)

	// Add members to manager
	prevLevel := ""
	for _, l := range levels {

		// Create a new member
		err := testBu.RelationManager().AddMembers([]*bursary.MemberEntry{
			l,
		}, prevLevel)
		if !assert.Nil(t, err) {
			break
		}

		prevLevel = l.ID
	}

	// Check members
	for _, l := range levels {

		// Check if member exists
		m, err := testBu.RelationManager().GetMember(l.ID)
		if assert.Nil(t, err) {
			assert.Equal(t, l.ID, m.ID)
		}
	}
}

func Test_RelationManagerPostgres_DeleteMembers(t *testing.T) {

	defer uninit()

	var levels []*bursary.MemberEntry

	// Preparing members
	me := bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 1.0,
		Share:      0,
	}
	levels = append(levels, me)

	me = bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 0.7,
		Share:      0.7,
	}
	levels = append(levels, me)

	me = bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 0.5,
		Share:      0.3,
	}
	levels = append(levels, me)

	// Add members to manager
	prevLevel := ""
	for _, l := range levels {

		// Create a new member
		err := testBu.RelationManager().AddMembers([]*bursary.MemberEntry{
			l,
		}, prevLevel)
		if !assert.Nil(t, err) {
			break
		}

		prevLevel = l.ID
	}

	// Delete members
	targetMembers := make([]string, 0)
	for _, l := range levels {
		targetMembers = append(targetMembers, l.ID)
	}

	err := testBu.RelationManager().DeleteMembers(targetMembers)
	if !assert.Nil(t, err) {
		return
	}

	// Check members
	for _, l := range levels {

		// Check if member exists
		_, err := testBu.RelationManager().GetMember(l.ID)
		assert.Error(t, err, bursary.ErrMemberNotFound)
	}
}

func Test_RelationManagerPostgres_ListMembers(t *testing.T) {

	defer uninit()

	var members []*bursary.MemberEntry

	// Preparing members
	me := bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 1.0,
		Share:      0,
	}
	members = append(members, me)

	me = bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 0.7,
		Share:      0.7,
	}
	members = append(members, me)

	me = bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 0.5,
		Share:      0.3,
	}
	members = append(members, me)

	// Add members to manager
	err := testBu.RelationManager().AddMembers(members, "")
	if !assert.Nil(t, err) {
		return
	}

	// General list
	cond := bursary.NewCondition()
	ms, err := testBu.RelationManager().ListMembers("", cond)
	if !assert.Nil(t, err) {
		return
	}

	assert.Len(t, ms, 3)

	// Pagination (Limit=1)
	cond.Limit = 1
	ms, err = testBu.RelationManager().ListMembers("", cond)
	if !assert.Nil(t, err) {
		return
	}

	assert.Len(t, ms, 1)

	// Pagination (Limit=1, Page=2)
	cond.Page = 2
	ms, err = testBu.RelationManager().ListMembers("", cond)
	if !assert.Nil(t, err) {
		return
	}

	assert.Len(t, ms, 1)
}

func Test_RelationManagerPostgres_ChangePath(t *testing.T) {

	defer uninit()

	// Preparing members
	me := bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 1.0,
		Share:      0,
	}

	// Create a new member
	err := testBu.RelationManager().AddMembers([]*bursary.MemberEntry{
		me,
	}, "")
	if !assert.Nil(t, err) {
		return
	}

	// Change path
	err = testBu.RelationManager().ChangePath(me.ID, []string{"test1", "test2"})
	if !assert.Nil(t, err) {
		return
	}

	// Check path
	paths, err := testBu.RelationManager().GetPath(me.ID)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, "test1", paths[0])
	assert.Equal(t, "test2", paths[1])
}

func Test_RelationManagerPostgres_MoveMembers(t *testing.T) {

	defer uninit()

	var levels []*bursary.MemberEntry

	// Preparing members
	me := bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 1.0,
		Share:      0,
	}
	levels = append(levels, me)

	me = bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 0.7,
		Share:      0.7,
	}
	levels = append(levels, me)

	me = bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 0.5,
		Share:      0.3,
	}
	levels = append(levels, me)

	// Add members to manager
	prevLevel := ""
	for _, l := range levels {

		// Create a new member
		err := testBu.RelationManager().AddMembers([]*bursary.MemberEntry{
			l,
		}, prevLevel)
		if !assert.Nil(t, err) {
			break
		}

		prevLevel = l.ID
	}

	// Move members to root
	err := testBu.RelationManager().MoveMembers([]string{levels[1].ID}, "")
	if !assert.Nil(t, err) {
		return
	}

	// Check path
	m, err := testBu.RelationManager().GetMember(levels[1].ID)
	assert.Nil(t, err)
	assert.Equal(t, RootNode, m.Upstream)
	assert.Len(t, m.RelationPath, 0)

	// Check second level
	m, err = testBu.RelationManager().GetMember(levels[2].ID)
	assert.Nil(t, err)
	assert.Equal(t, levels[1].ID, m.Upstream)
	assert.Len(t, m.RelationPath, 1)
	assert.Equal(t, levels[1].ID, m.RelationPath[0])
}

func Test_RelationManagerPostgres_GetUpstreams(t *testing.T) {

	defer uninit()

	var levels []*bursary.MemberEntry

	// Preparing members
	me := bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 1.0,
		Share:      0,
	}
	levels = append(levels, me)

	me = bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 0.7,
		Share:      0.7,
	}
	levels = append(levels, me)

	me = bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 0.5,
		Share:      0.3,
	}
	levels = append(levels, me)

	// Add members to manager
	prevLevel := ""
	for _, l := range levels {

		// Create a new member
		err := testBu.RelationManager().AddMembers([]*bursary.MemberEntry{
			l,
		}, prevLevel)
		if !assert.Nil(t, err) {
			break
		}

		prevLevel = l.ID
	}

	// Check upstreams
	members, err := testBu.RelationManager().GetUpstreams(levels[2].ID)
	if !assert.Nil(t, err) {
		return
	}

	assert.Len(t, members, 2)
	assert.Equal(t, members[0].ID, levels[0].ID)
	assert.Equal(t, members[1].ID, levels[1].ID)
}

func Test_RelationManagerPostgres_UpdateChannelRule(t *testing.T) {

	defer uninit()

	// Preparing members
	me := bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 1.0,
		Share:      0,
	}

	// Create a new member
	err := testBu.RelationManager().AddMembers([]*bursary.MemberEntry{
		me,
	}, "")
	if !assert.Nil(t, err) {
		return
	}

	// Update rule
	err = testBu.RelationManager().UpdateChannelRule(me.ID, "default", &bursary.Rule{
		Commission: 0.5,
		Share:      33,
	})
	if !assert.Nil(t, err) {
		return
	}

	// Add a new rule
	err = testBu.RelationManager().UpdateChannelRule(me.ID, "new", &bursary.Rule{
		Commission: 0.99,
		Share:      99,
	})
	if !assert.Nil(t, err) {
		return
	}

	// Get new member info
	m, err := testBu.RelationManager().GetMember(me.ID)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, float64(0.5), m.ChannelRules["default"].Commission)
	assert.Equal(t, float64(33), m.ChannelRules["default"].Share)

	assert.Equal(t, float64(0.99), m.ChannelRules["new"].Commission)
	assert.Equal(t, float64(99), m.ChannelRules["new"].Share)
}

func Test_RelationManagerPostgres_RemoveChannelRule(t *testing.T) {

	defer uninit()

	// Preparing members
	me := bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 1.0,
		Share:      0,
	}

	// Create a new member
	err := testBu.RelationManager().AddMembers([]*bursary.MemberEntry{
		me,
	}, "")
	if !assert.Nil(t, err) {
		return
	}

	// Remove rule
	err = testBu.RelationManager().RemoveChannelRule(me.ID, "default")
	if !assert.Nil(t, err) {
		return
	}

	// Get new member info
	m, err := testBu.RelationManager().GetMember(me.ID)
	if !assert.Nil(t, err) {
		return
	}

	assert.Nil(t, m.ChannelRules["default"])
}

func Test_RelationManagerPostgres_RemoveChannel(t *testing.T) {

	defer uninit()

	var levels []*bursary.MemberEntry

	// Preparing members
	me := bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 1.0,
		Share:      0,
	}
	levels = append(levels, me)

	me = bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 0.7,
		Share:      0.7,
	}
	levels = append(levels, me)

	me = bursary.NewMemberEntry()
	me.ChannelRules["default"] = &bursary.Rule{
		Commission: 0.5,
		Share:      0.3,
	}
	levels = append(levels, me)

	// Add members to manager
	prevLevel := ""
	for _, l := range levels {

		// Create a new member
		err := testBu.RelationManager().AddMembers([]*bursary.MemberEntry{
			l,
		}, prevLevel)
		if !assert.Nil(t, err) {
			break
		}

		prevLevel = l.ID
	}

	// Remvoe specific channel
	err := testBu.RelationManager().RemoveChannel("default")
	if !assert.Nil(t, err) {
		return
	}

	// Check members
	for _, l := range levels {

		// Check if member exists
		m, err := testBu.RelationManager().GetMember(l.ID)
		if !assert.Nil(t, err) {
			continue
		}

		assert.Nil(t, m.ChannelRules["default"])
	}
}
