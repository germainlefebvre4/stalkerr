package settings

import (
	"reflect"
	"testing"

	"github.com/glefebvre/stalkeer/internal/config"
)

// TestEveryConfigScalarFieldIsAccountedFor reflects over config.Config and
// asserts every scalar leaf field (string/int/int64/float64/bool) is either
// in the overridable registry, on the bootstrap list, or on the explicit
// exclusion list - so a future config.Config field added without updating
// this package fails a test instead of silently staying file/env-only. A
// slice-typed field (m3u.sources, filter.*.include_patterns/exclude_patterns)
// is list-shaped and is intentionally not visited: those are covered by
// their own dedicated mechanisms (m3u-source-overrides, filter-override-policy).
func TestEveryConfigScalarFieldIsAccountedFor(t *testing.T) {
	var unaccounted []string

	walkConfigFields(reflect.TypeOf(config.Config{}), "", func(dottedKey string) {
		if _, ok := FieldByKey(dottedKey); ok {
			return
		}
		if IsBootstrapKey(dottedKey) {
			return
		}
		if IsExcludedKey(dottedKey) {
			return
		}
		unaccounted = append(unaccounted, dottedKey)
	})

	if len(unaccounted) > 0 {
		t.Errorf("config.Config field(s) not registered as overridable, bootstrap, or explicitly excluded: %v", unaccounted)
	}
}

// walkConfigFields recursively visits every field of a config.Config (sub)struct
// t, calling fn with the dotted mapstructure path of each scalar leaf field
// (string/int/int64/float64/bool). Struct fields are recursed into; slice,
// map, and other field kinds are skipped, since they are list-shaped
// configuration covered by their own dedicated override mechanism.
func walkConfigFields(t reflect.Type, prefix string, fn func(dottedKey string)) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("mapstructure")
		if tag == "" || tag == "-" {
			continue
		}

		key := tag
		if prefix != "" {
			key = prefix + "." + tag
		}

		switch field.Type.Kind() {
		case reflect.Struct:
			walkConfigFields(field.Type, key, fn)
		case reflect.String, reflect.Int, reflect.Int64, reflect.Float64, reflect.Bool:
			fn(key)
		default:
			// Slices (e.g. m3u.sources, filter include/exclude patterns),
			// maps, etc. - list-shaped, not this registry's concern.
		}
	}
}
