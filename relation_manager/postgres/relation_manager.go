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

const RootNode = "00000000-0000-0000-0000-000000000000"

type Opt func(*RelationManagerPostgres)

type RelationManagerPostgres struct {
	db        *sqlx.DB
	tableName string
}

func NewRelationManagerPostgres(opts ...Opt) *RelationManagerPostgres {
	rm := &RelationManagerPostgres{}

	for _, opt := range opts {
		opt(rm)
	}

	if len(rm.tableName) == 0 {
		rm.tableName = "relationships"
	}

	return rm
}

func WithDb(db *sqlx.DB) Opt {
	return func(rm *RelationManagerPostgres) {
		rm.db = db
	}
}

func WithTableName(tableName string) Opt {
	return func(rm *RelationManagerPostgres) {
		rm.tableName = tableName
	}
}

func (rm *RelationManagerPostgres) Init() error {

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

func (rm *RelationManagerPostgres) Close() error {
	return rm.db.Close()
}

func (rm *RelationManagerPostgres) GetPath(mid string) ([]string, error) {

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

func (rm *RelationManagerPostgres) ChangePathByUpstream(upstream string, newPath []string) error {

	if len(upstream) == 0 {
		upstream = RootNode
	}

	cmd := fmt.Sprintf(`UPDATE %s SET relation_path = $1 WHERE upstream = $2 RETURNING id`, rm.tableName)
	rows, err := rm.db.Queryx(cmd, pq.StringArray(newPath), upstream)

	// Update downstreams
	record := &MemberRecord{}
	for rows.Next() {
		err := rows.StructScan(&record)
		if err != nil {
			return err
		}

		curPath := append(newPath, record.ID)
		err = rm.ChangePathByUpstream(record.ID, curPath)
		if err != nil {
			return err
		}
	}

	return err
}

func (rm *RelationManagerPostgres) ChangePath(mid string, newPath []string) error {

	cmd := fmt.Sprintf(`UPDATE %s SET relation_path = $1 WHERE id = $2`, rm.tableName)
	_, err := rm.db.Exec(cmd, pq.StringArray(newPath), mid)
	if err != nil {
		return err
	}

	// Update downstreams
	curPath := append(newPath, mid)
	return rm.ChangePathByUpstream(mid, curPath)
}

func (rm *RelationManagerPostgres) GetMember(mid string) (*bursary.Member, error) {

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

func (rm *RelationManagerPostgres) AddMembers(members []*bursary.MemberEntry, upstream string) error {

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
		upstream = RootNode
	}

	// Current timestamp
	ts := time.Now()

	// Preparing records
	records := make([]*MemberRecord, 0)
	for _, me := range members {

		m := &MemberRecord{
			ID:           me.ID,
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

func (rm *RelationManagerPostgres) MoveMembers(mids []string, upstream string) error {

	rp, err := rm.GetPath(upstream)
	if err != nil {
		return bursary.ErrUpstreamNotFound
	}

	if len(upstream) == 0 {
		upstream = RootNode
	}

	// update members
	cmd := fmt.Sprintf(`UPDATE %s SET upstream = $1, relation_path = $2 WHERE id = ANY ($3)`, rm.tableName)
	_, err = rm.db.Exec(cmd, upstream, pq.StringArray(rp), pq.Array(mids))
	if err != nil {
		return err
	}

	// update downstreams
	for _, mid := range mids {
		curPath := append(rp, mid)
		rm.ChangePathByUpstream(mid, curPath)
	}

	return nil
}

func (rm *RelationManagerPostgres) DeleteMembers(mids []string) error {

	cmd := fmt.Sprintf(`DELETE FROM %s WHERE id = ANY ($1)`, rm.tableName)
	_, err := rm.db.Exec(cmd, pq.Array(mids))

	return err
}

func (rm *RelationManagerPostgres) GetUpstreams(mid string) ([]*bursary.Member, error) {

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

func (rm *RelationManagerPostgres) ListMembers(upstream string, cond *bursary.Condition) ([]*bursary.Member, error) {

	if cond == nil {
		cond = bursary.NewCondition()
	}

	if len(upstream) == 0 {
		upstream = RootNode
	}

	members := make([]*bursary.Member, 0)

	offset := (cond.Page - 1) * cond.Limit

	cmd := fmt.Sprintf(`SELECT * FROM %s WHERE upstream = $1 OFFSET $2 LIMIT $3`, rm.tableName)
	rows, err := rm.db.Queryx(cmd, upstream, offset, cond.Limit)
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

func (rm *RelationManagerPostgres) UpdateChannelRule(mid string, channel string, rule *bursary.Rule) error {

	if rule == nil {
		return nil
	}

	ruleData, _ := json.Marshal(rule)

	cmd := fmt.Sprintf(`UPDATE %s SET channel_rules = jsonb_set(channel_rules, '{%s}', $1::jsonb) WHERE id = $2`, rm.tableName, channel)
	_, err := rm.db.Exec(cmd, ruleData, mid)

	return err
}

func (rm *RelationManagerPostgres) RemoveChannelRule(mid string, channel string) error {

	cmd := fmt.Sprintf(`UPDATE %s SET channel_rules = channel_rules - $1 WHERE id = $2`, rm.tableName)
	_, err := rm.db.Exec(cmd, channel, mid)

	return err
}

func (rm *RelationManagerPostgres) RemoveChannel(channel string) error {

	cmd := fmt.Sprintf(`UPDATE %s SET channel_rules = channel_rules - $1`, rm.tableName)
	_, err := rm.db.Exec(cmd, channel)

	return err
}
