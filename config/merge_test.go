package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMergeMaps validates the mergeMaps function which recursively merges
// configuration maps with src values overriding dst values.
func TestMergeMaps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dst  map[string]any
		src  map[string]any
		want map[string]any
	}{
		// Basic merging tests
		{
			name: "empty src - dst unchanged",
			dst:  map[string]any{"key1": "value1", "key2": 42},
			src:  map[string]any{},
			want: map[string]any{"key1": "value1", "key2": 42},
		},
		{
			name: "empty dst - src values added",
			dst:  map[string]any{},
			src:  map[string]any{"key1": "value1", "key2": 42},
			want: map[string]any{"key1": "value1", "key2": 42},
		},
		{
			name: "both empty - remains empty",
			dst:  map[string]any{},
			src:  map[string]any{},
			want: map[string]any{},
		},

		// Simple key-value merging
		{
			name: "non-overlapping keys - all retained",
			dst:  map[string]any{"key1": "value1"},
			src:  map[string]any{"key2": "value2"},
			want: map[string]any{"key1": "value1", "key2": "value2"},
		},
		{
			name: "overlapping keys - src overrides dst",
			dst:  map[string]any{"key1": "old"},
			src:  map[string]any{"key1": "new"},
			want: map[string]any{"key1": "new"},
		},
		{
			name: "mixed overlapping and non-overlapping",
			dst:  map[string]any{"key1": "old", "key2": "keep"},
			src:  map[string]any{"key1": "new", "key3": "add"},
			want: map[string]any{"key1": "new", "key2": "keep", "key3": "add"},
		},

		// Value type overriding
		{
			name: "different types - src overrides",
			dst:  map[string]any{"key1": "string"},
			src:  map[string]any{"key1": 42},
			want: map[string]any{"key1": 42},
		},
		{
			name: "override string with int",
			dst:  map[string]any{"port": "8080"},
			src:  map[string]any{"port": 8080},
			want: map[string]any{"port": 8080},
		},
		{
			name: "override int with bool",
			dst:  map[string]any{"enabled": 1},
			src:  map[string]any{"enabled": true},
			want: map[string]any{"enabled": true},
		},
		{
			name: "override primitive with slice",
			dst:  map[string]any{"items": "single"},
			src:  map[string]any{"items": []string{"a", "b", "c"}},
			want: map[string]any{"items": []string{"a", "b", "c"}},
		},
		{
			name: "override primitive with map",
			dst:  map[string]any{"config": "simple"},
			src:  map[string]any{"config": map[string]any{"nested": "value"}},
			want: map[string]any{"config": map[string]any{"nested": "value"}},
		},

		// Nested map merging (recursive)
		{
			name: "nested maps - merge recursively",
			dst: map[string]any{
				"server": map[string]any{
					"host": "localhost",
					"port": 8080,
				},
			},
			src: map[string]any{
				"server": map[string]any{
					"port": 9090,
				},
			},
			want: map[string]any{
				"server": map[string]any{
					"host": "localhost",
					"port": 9090,
				},
			},
		},
		{
			name: "nested maps - add new nested keys",
			dst: map[string]any{
				"server": map[string]any{
					"host": "localhost",
				},
			},
			src: map[string]any{
				"server": map[string]any{
					"port": 9090,
				},
			},
			want: map[string]any{
				"server": map[string]any{
					"host": "localhost",
					"port": 9090,
				},
			},
		},
		{
			name: "nested maps - multiple levels",
			dst: map[string]any{
				"app": map[string]any{
					"server": map[string]any{
						"host": "localhost",
					},
				},
			},
			src: map[string]any{
				"app": map[string]any{
					"server": map[string]any{
						"port": 9090,
					},
				},
			},
			want: map[string]any{
				"app": map[string]any{
					"server": map[string]any{
						"host": "localhost",
						"port": 9090,
					},
				},
			},
		},
		{
			name: "deeply nested maps - 4 levels deep",
			dst: map[string]any{
				"level1": map[string]any{
					"level2": map[string]any{
						"level3": map[string]any{
							"level4": "old",
						},
					},
				},
			},
			src: map[string]any{
				"level1": map[string]any{
					"level2": map[string]any{
						"level3": map[string]any{
							"level4": "new",
						},
					},
				},
			},
			want: map[string]any{
				"level1": map[string]any{
					"level2": map[string]any{
						"level3": map[string]any{
							"level4": "new",
						},
					},
				},
			},
		},

		// Non-map to map replacement (not recursive merge)
		{
			name: "non-map dst with map src - replace entirely",
			dst: map[string]any{
				"server": "simple-string",
			},
			src: map[string]any{
				"server": map[string]any{
					"host": "localhost",
					"port": 8080,
				},
			},
			want: map[string]any{
				"server": map[string]any{
					"host": "localhost",
					"port": 8080,
				},
			},
		},
		{
			name: "map dst with non-map src - replace entirely",
			dst: map[string]any{
				"server": map[string]any{
					"host": "localhost",
					"port": 8080,
				},
			},
			src: map[string]any{
				"server": "simple-string",
			},
			want: map[string]any{
				"server": "simple-string",
			},
		},

		// Complex realistic scenarios
		{
			name: "realistic config merge - file + env",
			dst: map[string]any{
				"server": map[string]any{
					"host": "0.0.0.0",
					"port": 8080,
					"timeout": 30,
				},
				"database": map[string]any{
					"host": "localhost",
					"port": 5432,
				},
				"debug": false,
			},
			src: map[string]any{
				"server": map[string]any{
					"port": 9090,
				},
				"debug": true,
			},
			want: map[string]any{
				"server": map[string]any{
					"host": "0.0.0.0",
					"port": 9090,
					"timeout": 30,
				},
				"database": map[string]any{
					"host": "localhost",
					"port": 5432,
				},
				"debug": true,
			},
		},
		{
			name: "realistic multi-source merge",
			dst: map[string]any{
				"app": map[string]any{
					"name": "myapp",
					"version": "1.0.0",
				},
			},
			src: map[string]any{
				"app": map[string]any{
					"version": "2.0.0",
				},
				"features": map[string]any{
					"auth": true,
					"logging": map[string]any{
						"level": "info",
					},
				},
			},
			want: map[string]any{
				"app": map[string]any{
					"name": "myapp",
					"version": "2.0.0",
				},
				"features": map[string]any{
					"auth": true,
					"logging": map[string]any{
						"level": "info",
					},
				},
			},
		},

		// Various data types
		{
			name: "merge with various primitive types",
			dst: map[string]any{
				"string": "value",
				"int": 42,
				"float": 3.14,
				"bool": true,
			},
			src: map[string]any{
				"int": 100,
				"new_key": "new_value",
			},
			want: map[string]any{
				"string": "value",
				"int": 100,
				"float": 3.14,
				"bool": true,
				"new_key": "new_value",
			},
		},
		{
			name: "merge with slices and arrays",
			dst: map[string]any{
				"hosts": []string{"host1", "host2"},
			},
			src: map[string]any{
				"hosts": []string{"host3", "host4"},
			},
			want: map[string]any{
				"hosts": []string{"host3", "host4"},
			},
		},
		{
			name: "merge with nil values",
			dst: map[string]any{
				"key1": "value1",
				"key2": nil,
			},
			src: map[string]any{
				"key2": "now-set",
				"key3": nil,
			},
			want: map[string]any{
				"key1": "value1",
				"key2": "now-set",
				"key3": nil,
			},
		},

		// Mixed nested and non-nested
		{
			name: "complex mixed structure",
			dst: map[string]any{
				"simple": "value",
				"nested": map[string]any{
					"inner1": "value1",
					"inner2": map[string]any{
						"deep": "value",
					},
				},
				"list": []int{1, 2, 3},
			},
			src: map[string]any{
				"simple": "updated",
				"nested": map[string]any{
					"inner2": map[string]any{
						"deep": "updated",
					},
					"inner3": "new",
				},
				"list": []int{4, 5},
			},
			want: map[string]any{
				"simple": "updated",
				"nested": map[string]any{
					"inner1": "value1",
					"inner2": map[string]any{
						"deep": "updated",
					},
					"inner3": "new",
				},
				"list": []int{4, 5},
			},
		},

		// Edge cases with empty nested maps
		{
			name: "empty nested map in src",
			dst: map[string]any{
				"server": map[string]any{
					"host": "localhost",
				},
			},
			src: map[string]any{
				"server": map[string]any{},
			},
			want: map[string]any{
				"server": map[string]any{
					"host": "localhost",
				},
			},
		},
		{
			name: "empty nested map in dst",
			dst: map[string]any{
				"server": map[string]any{},
			},
			src: map[string]any{
				"server": map[string]any{
					"host": "localhost",
				},
			},
			want: map[string]any{
				"server": map[string]any{
					"host": "localhost",
				},
			},
		},

		// Multiple nested maps at same level
		{
			name: "multiple sibling nested maps",
			dst: map[string]any{
				"server": map[string]any{
					"host": "localhost",
				},
				"database": map[string]any{
					"host": "localhost",
				},
			},
			src: map[string]any{
				"server": map[string]any{
					"port": 8080,
				},
				"database": map[string]any{
					"port": 5432,
				},
			},
			want: map[string]any{
				"server": map[string]any{
					"host": "localhost",
					"port": 8080,
				},
				"database": map[string]any{
					"host": "localhost",
					"port": 5432,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a copy of dst for testing to avoid mutation affecting test setup
			dstCopy := copyMap(tt.dst)

			// Act
			mergeMaps(dstCopy, tt.src)

			// Assert
			assert.Equal(t, tt.want, dstCopy, "merged map should match expected result")
		})
	}
}

// TestMergeMapsInPlace verifies that mergeMaps mutates the dst map in place
// and that the original dst is modified.
func TestMergeMapsInPlace(t *testing.T) {
	t.Parallel()

	t.Run("dst is modified in place", func(t *testing.T) {
		t.Parallel()

		dst := map[string]any{
			"key1": "value1",
		}
		src := map[string]any{
			"key2": "value2",
		}

		// Keep pointer to the map to verify it's the same instance
		dstPtr := &dst

		// Act
		mergeMaps(dst, src)

		// Assert - same map instance is modified
		assert.Same(t, &dst, dstPtr)
		assert.Equal(t, "value1", dst["key1"])
		assert.Equal(t, "value2", dst["key2"])
	})

	t.Run("nested maps are modified in place", func(t *testing.T) {
		t.Parallel()

		nestedMap := map[string]any{"host": "localhost"}
		dst := map[string]any{
			"server": nestedMap,
		}
		src := map[string]any{
			"server": map[string]any{
				"port": 8080,
			},
		}

		// Keep pointer to verify it's the same instance
		nestedMapPtr := &nestedMap

		// Act
		mergeMaps(dst, src)

		// Assert - same nested map instance is modified
		assert.Same(t, &nestedMap, nestedMapPtr)
		assert.Equal(t, "localhost", nestedMap["host"])
		assert.Equal(t, 8080, nestedMap["port"])
		// Also verify dst["server"] is the same map
		assert.Equal(t, nestedMap, dst["server"])
	})
}

// TestMergeMapsNilHandling tests how mergeMaps handles nil values.
func TestMergeMapsNilHandling(t *testing.T) {
	t.Parallel()

	t.Run("nil value in src - sets key to nil", func(t *testing.T) {
		t.Parallel()

		dst := map[string]any{
			"key1": "value1",
		}
		src := map[string]any{
			"key2": nil,
		}

		mergeMaps(dst, src)

		assert.Equal(t, "value1", dst["key1"])
		assert.Nil(t, dst["key2"])
	})

	t.Run("override non-nil with nil", func(t *testing.T) {
		t.Parallel()

		dst := map[string]any{
			"key1": "value1",
		}
		src := map[string]any{
			"key1": nil,
		}

		mergeMaps(dst, src)

		assert.Nil(t, dst["key1"])
	})

	t.Run("override nil with non-nil", func(t *testing.T) {
		t.Parallel()

		dst := map[string]any{
			"key1": nil,
		}
		src := map[string]any{
			"key1": "value1",
		}

		mergeMaps(dst, src)

		assert.Equal(t, "value1", dst["key1"])
	})

	t.Run("nil in nested map value", func(t *testing.T) {
		t.Parallel()

		dst := map[string]any{
			"config": map[string]any{
				"key1": "value1",
			},
		}
		src := map[string]any{
			"config": map[string]any{
				"key2": nil,
			},
		}

		mergeMaps(dst, src)

		nestedMap := dst["config"].(map[string]any)
		assert.Equal(t, "value1", nestedMap["key1"])
		assert.Nil(t, nestedMap["key2"])
	})
}

// TestMergeMapsMultipleSources simulates merging from multiple sources
// in precedence order, which is how Manager.Reload uses this function.
func TestMergeMapsMultipleSources(t *testing.T) {
	t.Parallel()

	t.Run("three source merge - later overrides earlier", func(t *testing.T) {
		t.Parallel()

		// Simulate file source
		merged := map[string]any{
			"server": map[string]any{
				"host": "0.0.0.0",
				"port": 8080,
			},
			"debug": false,
		}

		// Simulate env source
		envSource := map[string]any{
			"server": map[string]any{
				"port": 9090,
			},
		}
		mergeMaps(merged, envSource)

		// Simulate CLI source
		cliSource := map[string]any{
			"debug": true,
		}
		mergeMaps(merged, cliSource)

		// Assert final state
		assert.Equal(t, map[string]any{
			"server": map[string]any{
				"host": "0.0.0.0",
				"port": 9090,
			},
			"debug": true,
		}, merged)
	})

	t.Run("sequential merges maintain precedence", func(t *testing.T) {
		t.Parallel()

		result := map[string]any{}

		// Source 1: defaults
		mergeMaps(result, map[string]any{
			"timeout": 30,
			"retries": 3,
		})

		// Source 2: config file
		mergeMaps(result, map[string]any{
			"timeout": 60,
		})

		// Source 3: environment
		mergeMaps(result, map[string]any{
			"retries": 5,
		})

		assert.Equal(t, 60, result["timeout"], "timeout should be from config file")
		assert.Equal(t, 5, result["retries"], "retries should be from environment")
	})
}

// TestMergeMapsDeepCopyBehavior verifies the behavior regarding deep vs shallow copies.
func TestMergeMapsDeepCopyBehavior(t *testing.T) {
	t.Parallel()

	t.Run("modifying src after merge does not affect dst", func(t *testing.T) {
		t.Parallel()

		dst := map[string]any{}
		src := map[string]any{
			"key": "value",
		}

		mergeMaps(dst, src)

		// Modify src
		src["key"] = "modified"

		// dst should not be affected (primitive values are copied)
		assert.Equal(t, "value", dst["key"])
	})

	t.Run("nested maps share reference - modifications affect both", func(t *testing.T) {
		t.Parallel()

		nestedMap := map[string]any{"inner": "value"}
		dst := map[string]any{}
		src := map[string]any{
			"config": nestedMap,
		}

		mergeMaps(dst, src)

		// Modify the nested map through original reference
		nestedMap["inner"] = "modified"

		// dst's nested map is the same reference
		dstNested := dst["config"].(map[string]any)
		assert.Equal(t, "modified", dstNested["inner"])
	})
}

// BenchmarkMergeMaps measures the performance of mergeMaps with various scenarios.
func BenchmarkMergeMaps(b *testing.B) {
	benchmarks := []struct {
		name string
		dst  map[string]any
		src  map[string]any
	}{
		{
			name: "flat map - 5 keys",
			dst:  map[string]any{"k1": "v1", "k2": "v2", "k3": "v3"},
			src:  map[string]any{"k4": "v4", "k5": "v5"},
		},
		{
			name: "flat map - 20 keys",
			dst: map[string]any{
				"k1": "v1", "k2": "v2", "k3": "v3", "k4": "v4", "k5": "v5",
				"k6": "v6", "k7": "v7", "k8": "v8", "k9": "v9", "k10": "v10",
			},
			src: map[string]any{
				"k11": "v11", "k12": "v12", "k13": "v13", "k14": "v14", "k15": "v15",
				"k16": "v16", "k17": "v17", "k18": "v18", "k19": "v19", "k20": "v20",
			},
		},
		{
			name: "nested maps - 2 levels",
			dst: map[string]any{
				"server": map[string]any{"host": "localhost", "port": 8080},
				"db":     map[string]any{"host": "localhost", "port": 5432},
			},
			src: map[string]any{
				"server": map[string]any{"port": 9090},
				"db":     map[string]any{"user": "admin"},
			},
		},
		{
			name: "nested maps - 4 levels deep",
			dst: map[string]any{
				"l1": map[string]any{
					"l2": map[string]any{
						"l3": map[string]any{
							"l4": "value",
						},
					},
				},
			},
			src: map[string]any{
				"l1": map[string]any{
					"l2": map[string]any{
						"l3": map[string]any{
							"l4": "new",
						},
					},
				},
			},
		},
		{
			name: "mixed nested and flat",
			dst: map[string]any{
				"simple1": "value1",
				"simple2": "value2",
				"nested": map[string]any{
					"inner1": "value",
					"inner2": map[string]any{"deep": "value"},
				},
			},
			src: map[string]any{
				"simple3": "value3",
				"nested": map[string]any{
					"inner3": "new",
				},
			},
		},
		{
			name: "all overrides - no merges",
			dst:  map[string]any{"k1": "old1", "k2": "old2", "k3": "old3"},
			src:  map[string]any{"k1": "new1", "k2": "new2", "k3": "new3"},
		},
		{
			name: "no overrides - all new keys",
			dst:  map[string]any{"k1": "v1", "k2": "v2"},
			src:  map[string]any{"k3": "v3", "k4": "v4"},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				// Copy maps for each iteration to avoid cumulative state
				dst := copyMap(bm.dst)
				src := copyMap(bm.src)
				mergeMaps(dst, src)
			}
		})
	}
}

// copyMap creates a shallow copy of a map for testing purposes.
// This helper ensures test isolation by preventing tests from modifying shared data.
func copyMap(m map[string]any) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		if nestedMap, ok := v.(map[string]any); ok {
			result[k] = copyMap(nestedMap)
		} else {
			result[k] = v
		}
	}
	return result
}
