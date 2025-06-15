package db

import (
	"encoding/json"
	"fmt"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/logs"
)

func ToJSONB(src any, v any) error {
	var source []byte
	switch t := src.(type) {
	case string:
		source = []byte(t)
	case []byte:
		source = t
	case nil:
		return nil
	default:
		return fmt.Errorf("incompatible data type %T", src)
	}
	if len(source) == 0 {
		return nil
	}
	return json.Unmarshal(source, v)
}

func ToBytes(src map[string]any) ([]byte, error) {
	data, err := json.Marshal(src)
	if err != nil {
		logs.Logger.Error().Err(err).Msg("Failed to marshal data to JSON")
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	return data, nil
}