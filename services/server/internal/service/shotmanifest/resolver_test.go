package shotmanifest

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestResolveStateInheritsWhenNoExplicitChange(t *testing.T) {
	previous := `{
		"characters": {
			"char_hansanhe": {
				"look": {"outerwear":"half_removed","hair":"tied_messy"},
				"injury": {"forehead_wound":"present"}
			}
		},
		"props": {"wooden_spoon":{"owner":"char_hansanhe","location":"waist"}}
	}`
	resolved, err := ResolveState(previous, `{}`)
	if err != nil {
		t.Fatalf("ResolveState() error = %v", err)
	}
	if !jsonObjectsEqual(t, resolved, previous) {
		t.Fatalf("resolved = %s, want inherited %s", resolved, previous)
	}
}

func TestResolveStateOnlyOverridesExplicitNestedChange(t *testing.T) {
	previous := `{
		"characters": {
			"char_hansanhe": {
				"look": {"outerwear":"half_removed","hair":"tied_messy"},
				"injury": {"forehead_wound":"present"}
			}
		},
		"scene": {"id":"cafeteria","zone":"kitchen_passage"}
	}`
	changes := `{
		"characters": {
			"char_hansanhe": {
				"look": {"outerwear":"restored"}
			}
		}
	}`
	want := `{
		"characters": {
			"char_hansanhe": {
				"look": {"outerwear":"restored","hair":"tied_messy"},
				"injury": {"forehead_wound":"present"}
			}
		},
		"scene": {"id":"cafeteria","zone":"kitchen_passage"}
	}`
	resolved, err := ResolveState(previous, changes)
	if err != nil {
		t.Fatalf("ResolveState() error = %v", err)
	}
	if !jsonObjectsEqual(t, resolved, want) {
		t.Fatalf("resolved = %s, want %s", resolved, want)
	}
}

func TestResolveStateExplicitNullClearsStateWithoutResettingSiblings(t *testing.T) {
	previous := `{
		"characters": {
			"char_a": {
				"look": {"wardrobe":"wedding","hair":"updo"},
				"props": {"bouquet":"held"}
			}
		}
	}`
	changes := `{"characters":{"char_a":{"props":{"bouquet":null}}}}`
	want := `{
		"characters": {
			"char_a": {
				"look": {"wardrobe":"wedding","hair":"updo"},
				"props": {}
			}
		}
	}`
	resolved, err := ResolveState(previous, changes)
	if err != nil {
		t.Fatalf("ResolveState() error = %v", err)
	}
	if !jsonObjectsEqual(t, resolved, want) {
		t.Fatalf("resolved = %s, want %s", resolved, want)
	}
}

func TestResolveStateRejectsMalformedJSON(t *testing.T) {
	if _, err := ResolveState(`{"characters":`, `{}`); err == nil {
		t.Fatal("ResolveState(malformed previous) error = nil, want error")
	}
	if _, err := ResolveState(`{}`, `{"changes":`); err == nil {
		t.Fatal("ResolveState(malformed changes) error = nil, want error")
	}
}

func jsonObjectsEqual(t *testing.T, first string, second string) bool {
	t.Helper()
	var left any
	var right any
	if err := json.Unmarshal([]byte(first), &left); err != nil {
		t.Fatalf("decode first JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(second), &right); err != nil {
		t.Fatalf("decode second JSON: %v", err)
	}
	return reflect.DeepEqual(left, right)
}
