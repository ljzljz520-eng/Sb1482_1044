package store

import (
	"encoding/json"
	"fmt"
)

func encode(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode value: %w", err)
	}
	return data, nil
}

func decode(data []byte, target any) error {
	if len(data) == 0 {
		return fmt.Errorf("empty stored value")
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode value: %w", err)
	}
	return nil
}

func cloneBytes(data []byte) []byte {
	copyData := make([]byte, len(data))
	copy(copyData, data)
	return copyData
}
