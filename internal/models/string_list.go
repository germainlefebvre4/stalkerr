package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// StringList is a []string persisted as a JSON array in a TEXT column,
// portable across the Postgres and SQLite drivers used by this project.
type StringList []string

// GormDataType marks this as a plain leaf field (not a slice relationship)
// during schema parsing.
func (StringList) GormDataType() string {
	return "text"
}

// GormDBDataType ensures AutoMigrate creates a plain TEXT column on every
// supported driver, rather than inferring one from the underlying slice kind.
func (StringList) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "text"
}

// Value implements driver.Valuer.
func (s StringList) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	b, err := json.Marshal([]string(s))
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements sql.Scanner.
func (s *StringList) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported type for StringList.Scan: %T", value)
	}

	if len(b) == 0 {
		*s = nil
		return nil
	}

	var out []string
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	*s = out
	return nil
}
