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
var testRM *RelationManager
var testBu bursary.Bursary

func init() {

	// Connect to postgres server
	db, err := sqlx.Connect("postgres", "port=32768 user=postgres password=1qazXSW@ dbname=bursary sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}

	testDb = db

	rm := NewRelationManager(
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

func Test_RelationManager_AddMembers(t *testing.T) {

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

		prevLevel = l.Id
	}

	// Check members
	for _, l := range levels {

		// Check if member exists
		m, err := testBu.RelationManager().GetMember(l.Id)
		if assert.Nil(t, err) {
			assert.Equal(t, l.Id, m.Id)
		}
	}
}

func Test_RelationManager_DeleteMembers(t *testing.T) {

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

		prevLevel = l.Id
	}

	// Delete members
	targetMembers := make([]string, 0)
	for _, l := range levels {
		targetMembers = append(targetMembers, l.Id)
	}

	err := testBu.RelationManager().DeleteMembers(targetMembers)
	if !assert.Nil(t, err) {
		return
	}

	// Check members
	for _, l := range levels {

		// Check if member exists
		_, err := testBu.RelationManager().GetMember(l.Id)
		assert.Error(t, err, bursary.ErrMemberNotFound)
	}
}

func Test_RelationManager_GetUpstreams(t *testing.T) {

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

		prevLevel = l.Id
	}

	// Check upstreams
	members, err := testBu.RelationManager().GetUpstreams(levels[2].Id)
	if !assert.Nil(t, err) {
		return
	}

	assert.Len(t, members, 2)
	assert.Equal(t, members[0].Id, levels[0].Id)
	assert.Equal(t, members[1].Id, levels[1].Id)
}
