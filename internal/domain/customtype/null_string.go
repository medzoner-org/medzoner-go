package customtype

import (
	"database/sql/driver"
	"encoding/json"
)

// NullString represents a string that may be null.
// It implements sql.Scanner, driver.Valuer, and json.Marshaler/json.Unmarshaler.
type NullString struct {
	String string
	Valid  bool
}

// NewNullString creates a valid NullString with the given value.
func NewNullString(s string) NullString {
	return NullString{String: s, Valid: true}
}

// Scan implements the sql.Scanner interface.
func (ns *NullString) Scan(value interface{}) error {
	if value == nil {
		ns.String = ""
		ns.Valid = false
		return nil
	}
	switch v := value.(type) {
	case string:
		ns.String = v
	case []byte:
		ns.String = string(v)
	default:
		ns.String = ""
		ns.Valid = false
		return nil
	}
	ns.Valid = true
	return nil
}

// Value implements the driver.Valuer interface.
func (ns NullString) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.String, nil
}

// MarshalJSON implements the json.Marshaler interface.
func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(ns.String)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (ns *NullString) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == nil {
		ns.String = ""
		ns.Valid = false
		return nil
	}
	ns.String = *s
	ns.Valid = true
	return nil
}
