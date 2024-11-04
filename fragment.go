package main

import (
	"strings"
)

type Fragment struct {
	name string
	code string
	meta *CoreTable
}

func (f *Fragment) getMeta() *map[string]CoreType {
	return &f.meta.v
}

func (f *Fragment) setMeta(key string, value CoreType) {
	// allow setting nested keys ie. "foo.bar.baz"

	keys := strings.Split(key, ".")
	current := f.meta
	for i, k := range keys {
		if i == len(keys)-1 {
			// Last key, set the value
			current.v[k] = value
			return
		}
		// Intermediate keys, expect CoreTable
		if val, ok := current.v[k]; ok {
			if ct, ok := val.(*CoreTable); ok {
				current = ct
			} else {
				// Not a table, create a new one
				newTable := NewCoreTable(make(map[string]CoreType))
				current.v[k] = newTable
				current = newTable
			}
		} else {
			// Key does not exist, create a new table
			newTable := NewCoreTable(make(map[string]CoreType))
			current.v[k] = newTable
			current = newTable
		}
	}
}

// TODO: new fragment refactor add some evaluation functions and link to LFragment
