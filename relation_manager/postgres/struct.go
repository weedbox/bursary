package relation_manager_postgres

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

type Rule struct {
	Commission float64 `json:"commission"`
	Share      float64 `json:"share"`
}

type ChannelRules map[string]*Rule

func (cr ChannelRules) Value() (driver.Value, error) {
	return json.Marshal(cr)
}

func (cr *ChannelRules) Scan(src interface{}) error {

	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	var r ChannelRules
	err := json.Unmarshal(source, &r)
	if err != nil {
		return err
	}

	*cr = r

	return nil
}

type MemberRecord struct {
	ID           string         `db:"id"`
	ChannelRules ChannelRules   `db:"channel_rules"`
	RelationPath pq.StringArray `db:"relation_path"`
	Upstream     string         `db:"upstream"`
	CreatedAt    time.Time      `db:"created_at"`
}
