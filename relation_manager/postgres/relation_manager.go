package relation_manager_postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kulado/sqlxmigrate"
	"github.com/lib/pq"
	"github.com/weedbox/bursary"
)

type Opt func(*RelationManager)

type RelationManager struct {
	db        *sqlx.DB
	tableName string
}

func NewRelationManager(opts ...Opt) *RelationManager {
	rm := &RelationManager{}

	for _, opt := range opts {
		opt(rm)
	}

	if len(rm.tableName) == 0 {
		rm.tableName = "relationships"
	}

	return rm
}

func WithDb(db *sqlx.DB) Opt {
	return func(rm *RelationManager) {
		rm.db = db
	}
}

func WithTableName(tableName string) Opt {
	return func(rm *RelationManager) {
		rm.tableName = tableName
	}
}

func (rm *RelationManager) Init() error {

	// Initializing table
	m := sqlxmigrate.New(rm.db, sqlxmigrate.DefaultOptions, []*sqlxmigrate.Migration{
		{
			ID: "202306040726",
			Migrate: func(tx *sql.Tx) error {

				q := fmt.Sprintf(`CREATE TABLE "%s" (
						"id" UUID,
						"channel_rules" JSONB,
						"relation_path" TEXT[],
						"upstream" UUID,
						"created_at" timestamp with time zone,
						PRIMARY KEY ("id")
					)`, rm.tableName)

				_, err := tx.Exec(q)
				return err
			},
			Rollback: func(tx *sql.Tx) error {
				q := fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, rm.tableName)
				_, err := tx.Exec(q)
				return err
			},
		},
	})

	if err := m.Migrate(); err != nil {
		return err
	}

	return nil
}

func (rm *RelationManager) Close() error {
	return rm.db.Close()
}

func (rm *RelationManager) GetPath(mid string) ([]string, error) {

	if len(mid) == 0 {
		return []string{}, nil
	}

	cmd := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1`, rm.tableName)
	record := &MemberRecord{}
	err := rm.db.Get(record, cmd, mid)
	if err != nil {
		return []string{}, err
	}

	p := make([]string, 0)
	p = append(p, record.RelationPath...)
	p = append(p, mid)

	return p, nil
}

func (rm *RelationManager) ChangePath(mid string, newPath []string) error {

	cmd := fmt.Sprintf(`UPDATE %s SET relation_path = $1 WHERE id = $2`, rm.tableName)
	_, err := rm.db.Exec(cmd, pq.StringArray(newPath), mid)

	return err
}

func (rm *RelationManager) GetMember(mid string) (*bursary.Member, error) {

	if len(mid) == 0 {
		return nil, bursary.ErrMemberNotFound
	}

	cmd := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1`, rm.tableName)
	records := []MemberRecord{}
	err := rm.db.Select(&records, cmd, mid)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, bursary.ErrMemberNotFound
	}

	return records[0].ToMemberObject(), nil
}

func (rm *RelationManager) AddMembers(members []*bursary.MemberEntry, upstream string) error {

	if len(members) == 0 {
		return bursary.ErrMemberRequired
	}

	// Make sure that upstream exists
	rp, err := rm.GetPath(upstream)
	if err != nil {
		return bursary.ErrUpstreamNotFound
	}

	// Check upstream
	_, err = uuid.Parse(upstream)
	if err != nil {
		// For root node
		upstream = "00000000-0000-0000-0000-000000000000"
	}

	// Current timestamp
	ts := time.Now()

	// Preparing records
	records := make([]*MemberRecord, 0)
	for _, me := range members {

		m := &MemberRecord{
			Id:           me.Id,
			ChannelRules: make(map[string]*Rule),
			RelationPath: pq.StringArray(rp),
			Upstream:     upstream,
			CreatedAt:    ts,
		}

		for channel, cr := range me.ChannelRules {
			m.ChannelRules[channel] = &Rule{
				Commission: cr.Commission,
				Share:      cr.Share,
			}
		}

		records = append(records, m)
	}

	// Execute
	cmd := fmt.Sprintf(`INSERT INTO "%s" (
			id,
			channel_rules,
			relation_path,
			upstream,
			created_at
		) VALUES (
			:id,
			:channel_rules,
			:relation_path,
			:upstream,
			:created_at
		)`, rm.tableName)

	_, err = rm.db.NamedExec(cmd, records)

	return err
}

func (rm *RelationManager) MoveMembers(mids []string, upstream string) error {
	return nil
}

func (rm *RelationManager) DeleteMembers(mids []string) error {

	cmd := fmt.Sprintf(`DELETE FROM %s WHERE id = ANY ($1)`, rm.tableName)
	_, err := rm.db.Exec(cmd, pq.Array(mids))

	return err
}

func (rm *RelationManager) GetUpstreams(mid string) ([]*bursary.Member, error) {

	members := make([]*bursary.Member, 0)

	cmd := fmt.Sprintf(`SELECT * FROM %s WHERE id::text IN (
		SELECT unnest(relation_path) FROM %s WHERE id = $1
	)`, rm.tableName, rm.tableName)

	rows, err := rm.db.Queryx(cmd, mid)
	if err != nil {
		return members, err
	}

	record := &MemberRecord{}
	for rows.Next() {
		err := rows.StructScan(&record)
		if err != nil {
			return members, err
		}

		members = append(members, record.ToMemberObject())
	}

	return members, nil
}

func (rm *RelationManager) ListMembers(cond *bursary.Condition) ([]*bursary.Member, error) {
	members := make([]*bursary.Member, 0)
	return members, nil
}

func (rm *RelationManager) UpdateChannelRule(mid string, channel string, rule *bursary.Rule) error {

	if rule == nil {
		return nil
	}

	ruleData, _ := json.Marshal(rule)

	cmd := fmt.Sprintf(`UPDATE %s SET channel_rules = jsonb_set(channel_rules, '{%s}', $1::jsonb) WHERE id = $2`, rm.tableName, channel)
	_, err := rm.db.Exec(cmd, ruleData, mid)

	return err
}

func (rm *RelationManager) RemoveChannelRule(mid string, channel string) error {

	cmd := fmt.Sprintf(`UPDATE %s SET channel_rules = channel_rules - $1 WHERE id = $2`, rm.tableName)
	_, err := rm.db.Exec(cmd, channel, mid)

	return err
}

func (rm *RelationManager) RemoveChannel(channel string) error {

	cmd := fmt.Sprintf(`UPDATE %s SET channel_rules = channel_rules - $1`, rm.tableName)
	_, err := rm.db.Exec(cmd, channel)

	return err
}
