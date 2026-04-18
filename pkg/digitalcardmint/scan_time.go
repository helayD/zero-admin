package digitalcardmint

import (
	"database/sql"
	"database/sql/driver"
	"strings"
	"time"
)

type nullableTime struct {
	time  time.Time
	valid bool
}

var _ sql.Scanner = (*nullableTime)(nil)
var _ driver.Valuer = (*nullableTime)(nil)

func (n *nullableTime) Scan(value interface{}) error {
	switch typed := value.(type) {
	case nil:
		n.time = time.Time{}
		n.valid = false
		return nil
	case time.Time:
		n.time = typed
		n.valid = true
		return nil
	case []byte:
		return n.parseString(string(typed))
	case string:
		return n.parseString(typed)
	default:
		n.time = time.Time{}
		n.valid = false
		return nil
	}
}

func (n *nullableTime) parseString(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		n.time = time.Time{}
		n.valid = false
		return nil
	}

	layouts := []string{
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
		time.RFC3339Nano,
		time.RFC3339,
	}
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, trimmed, time.Local)
		if err == nil {
			n.time = parsed
			n.valid = true
			return nil
		}
	}

	n.time = time.Time{}
	n.valid = false
	return nil
}

func (n nullableTime) Value() (driver.Value, error) {
	if !n.valid || n.time.IsZero() {
		return nil, nil
	}
	return n.time, nil
}

func formatNullableTime(value nullableTime) string {
	if !value.valid || value.time.IsZero() {
		return ""
	}
	return value.time.Format("2006-01-02 15:04:05")
}
