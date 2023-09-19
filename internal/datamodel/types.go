package datamodel

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type JSON []map[string]interface{}

// Scan implements the Scanner interface.
func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := make([]map[string]interface{}, 0)
	err := json.Unmarshal(bytes, &result)
	*j = JSON(result)
	return err
}

// Value implements the driver Valuer interface.
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}

	return json.Marshal(&j)
}
