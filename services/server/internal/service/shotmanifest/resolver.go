package shotmanifest

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ResolveState applies explicit current-shot state changes on top of the
// previous resolved continuity state. Objects merge recursively, arrays/scalars
// replace atomically, and explicit null removes a key.
func ResolveState(previousJSON string, changesJSON string) (string, error) {
	previous, err := decodeStateObject(previousJSON)
	if err != nil {
		return "", fmt.Errorf("decoding previous shot state: %w", err)
	}
	changes, err := decodeStateObject(changesJSON)
	if err != nil {
		return "", fmt.Errorf("decoding shot state changes: %w", err)
	}
	resolved := cloneStateMap(previous)
	mergeStateMap(resolved, changes)
	encoded, err := json.Marshal(resolved)
	if err != nil {
		return "", fmt.Errorf("encoding resolved shot state: %w", err)
	}
	return string(encoded), nil
}

func decodeStateObject(raw string) (map[string]any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]any{}, nil
	}
	var value map[string]any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil, err
	}
	if value == nil {
		return map[string]any{}, nil
	}
	return value, nil
}

func mergeStateMap(target map[string]any, changes map[string]any) {
	for key, change := range changes {
		if change == nil {
			delete(target, key)
			continue
		}
		changeMap, changeIsMap := change.(map[string]any)
		if !changeIsMap {
			target[key] = cloneStateValue(change)
			continue
		}
		existingMap, existingIsMap := target[key].(map[string]any)
		if !existingIsMap {
			existingMap = map[string]any{}
		} else {
			existingMap = cloneStateMap(existingMap)
		}
		mergeStateMap(existingMap, changeMap)
		target[key] = existingMap
	}
}

func cloneStateMap(value map[string]any) map[string]any {
	result := make(map[string]any, len(value))
	for key, item := range value {
		result[key] = cloneStateValue(item)
	}
	return result
}

func cloneStateValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneStateMap(typed)
	case []any:
		result := make([]any, len(typed))
		for index, item := range typed {
			result[index] = cloneStateValue(item)
		}
		return result
	default:
		return typed
	}
}
