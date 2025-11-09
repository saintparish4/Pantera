package repository

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONB is a custom type for PostgreSQL JSONB columns
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONB value: not a byte slice")
	}

	result := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = result
	return nil
}

// ToMap converts JSONB to map[string]interface{}
func (j JSONB) ToMap() map[string]interface{} {
	if j == nil {
		return make(map[string]interface{})
	}
	return map[string]interface{}(j)
}

// FromMap creates JSONB from map[string]interface{}
func FromMap(m map[string]interface{}) JSONB {
	if m == nil {
		return make(JSONB)
	}
	return JSONB(m)
}
