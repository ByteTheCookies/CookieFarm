package database

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
)

type Cursor struct {
	Time int64 `json:"t"`
	ID   int64 `json:"i"`
}

func EncodeCursor(submitTime, id int64) string {
	b, err := json.Marshal(Cursor{Time: submitTime, ID: id})
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

func ParseCursor(s string) (sql.NullInt64, sql.NullInt64) {
	if s == "" {
		return sql.NullInt64{}, sql.NullInt64{}
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return sql.NullInt64{}, sql.NullInt64{}
	}
	var cur Cursor
	if err := json.Unmarshal(b, &cur); err != nil {
		return sql.NullInt64{}, sql.NullInt64{}
	}
	return sql.NullInt64{Int64: cur.Time, Valid: true},
		sql.NullInt64{Int64: cur.ID, Valid: true}
}
