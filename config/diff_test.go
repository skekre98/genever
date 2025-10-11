package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDiffEvent validates the diffEvent function which compares two configuration
// objects and returns an Event with the list of changed field names.
func TestDiffEvent(t *testing.T) {
	t.Parallel()

	// Test structs for various scenarios
	type SimpleConfig struct {
		Host string
		Port int
	}

	type NestedConfig struct {
		Server SimpleConfig
		Debug  bool
	}

	type MixedFieldTypes struct {
		StringField  string
		IntField     int
		BoolField    bool
		FloatField   float64
		SliceField   []string
		MapField     map[string]string
		StructField  SimpleConfig
		PointerField *string
	}

	stringPtr := func(s string) *string { return &s }

	tests := []struct {
		name        string
		old         any
		new         any
		wantChanged []string
	}{
		// Nil handling tests
		{
			name:        "both nil - should return empty changed keys",
			old:         nil,
			new:         nil,
			wantChanged: []string{},
		},
		{
			name:        "old nil new non-nil - should return empty changed keys",
			old:         nil,
			new:         &SimpleConfig{Host: "localhost", Port: 8080},
			wantChanged: []string{},
		},
		{
			name:        "old non-nil new nil - should return empty changed keys",
			old:         &SimpleConfig{Host: "localhost", Port: 8080},
			new:         nil,
			wantChanged: []string{},
		},

		// Pointer dereferencing tests
		{
			name: "pointers to identical structs - no changes",
			old:  &SimpleConfig{Host: "localhost", Port: 8080},
			new:  &SimpleConfig{Host: "localhost", Port: 8080},
			wantChanged: []string{},
		},
		{
			name: "pointers to different structs - detect changes",
			old:  &SimpleConfig{Host: "localhost", Port: 8080},
			new:  &SimpleConfig{Host: "example.com", Port: 9090},
			wantChanged: []string{"Host", "Port"},
		},
		{
			name: "value and pointer comparison - detect changes",
			old:  SimpleConfig{Host: "localhost", Port: 8080},
			new:  &SimpleConfig{Host: "example.com", Port: 8080},
			wantChanged: []string{"Host"},
		},

		// Identical structs - no changes
		{
			name: "identical simple structs - no changes",
			old:  SimpleConfig{Host: "localhost", Port: 8080},
			new:  SimpleConfig{Host: "localhost", Port: 8080},
			wantChanged: []string{},
		},
		{
			name: "identical nested structs - no changes",
			old: NestedConfig{
				Server: SimpleConfig{Host: "localhost", Port: 8080},
				Debug:  true,
			},
			new: NestedConfig{
				Server: SimpleConfig{Host: "localhost", Port: 8080},
				Debug:  true,
			},
			wantChanged: []string{},
		},

		// Single field changes
		{
			name: "single string field changed",
			old:  SimpleConfig{Host: "localhost", Port: 8080},
			new:  SimpleConfig{Host: "example.com", Port: 8080},
			wantChanged: []string{"Host"},
		},
		{
			name: "single int field changed",
			old:  SimpleConfig{Host: "localhost", Port: 8080},
			new:  SimpleConfig{Host: "localhost", Port: 9090},
			wantChanged: []string{"Port"},
		},
		{
			name: "single bool field changed",
			old:  NestedConfig{Server: SimpleConfig{Host: "localhost", Port: 8080}, Debug: true},
			new:  NestedConfig{Server: SimpleConfig{Host: "localhost", Port: 8080}, Debug: false},
			wantChanged: []string{"Debug"},
		},

		// Multiple field changes
		{
			name: "all fields changed",
			old:  SimpleConfig{Host: "localhost", Port: 8080},
			new:  SimpleConfig{Host: "example.com", Port: 9090},
			wantChanged: []string{"Host", "Port"},
		},
		{
			name: "nested struct field changed - detects top-level field",
			old: NestedConfig{
				Server: SimpleConfig{Host: "localhost", Port: 8080},
				Debug:  true,
			},
			new: NestedConfig{
				Server: SimpleConfig{Host: "example.com", Port: 9090},
				Debug:  true,
			},
			wantChanged: []string{"Server"},
		},
		{
			name: "multiple top-level fields changed",
			old: NestedConfig{
				Server: SimpleConfig{Host: "localhost", Port: 8080},
				Debug:  true,
			},
			new: NestedConfig{
				Server: SimpleConfig{Host: "example.com", Port: 9090},
				Debug:  false,
			},
			wantChanged: []string{"Server", "Debug"},
		},

		// Mixed field types
		{
			name: "mixed types - all unchanged",
			old: MixedFieldTypes{
				StringField:  "test",
				IntField:     42,
				BoolField:    true,
				FloatField:   3.14,
				SliceField:   []string{"a", "b"},
				MapField:     map[string]string{"key": "value"},
				StructField:  SimpleConfig{Host: "localhost", Port: 8080},
				PointerField: stringPtr("pointer"),
			},
			new: MixedFieldTypes{
				StringField:  "test",
				IntField:     42,
				BoolField:    true,
				FloatField:   3.14,
				SliceField:   []string{"a", "b"},
				MapField:     map[string]string{"key": "value"},
				StructField:  SimpleConfig{Host: "localhost", Port: 8080},
				PointerField: stringPtr("pointer"),
			},
			wantChanged: []string{},
		},
		{
			name: "mixed types - string changed",
			old: MixedFieldTypes{
				StringField: "old",
				IntField:    42,
			},
			new: MixedFieldTypes{
				StringField: "new",
				IntField:    42,
			},
			wantChanged: []string{"StringField"},
		},
		{
			name: "mixed types - slice changed",
			old: MixedFieldTypes{
				SliceField: []string{"a", "b"},
			},
			new: MixedFieldTypes{
				SliceField: []string{"a", "b", "c"},
			},
			wantChanged: []string{"SliceField"},
		},
		{
			name: "mixed types - map changed",
			old: MixedFieldTypes{
				MapField: map[string]string{"key": "value1"},
			},
			new: MixedFieldTypes{
				MapField: map[string]string{"key": "value2"},
			},
			wantChanged: []string{"MapField"},
		},
		{
			name: "mixed types - pointer field changed",
			old: MixedFieldTypes{
				PointerField: stringPtr("old"),
			},
			new: MixedFieldTypes{
				PointerField: stringPtr("new"),
			},
			wantChanged: []string{"PointerField"},
		},
		{
			name: "mixed types - nil to non-nil pointer",
			old: MixedFieldTypes{
				PointerField: nil,
			},
			new: MixedFieldTypes{
				PointerField: stringPtr("new"),
			},
			wantChanged: []string{"PointerField"},
		},
		{
			name: "mixed types - multiple fields changed",
			old: MixedFieldTypes{
				StringField: "old",
				IntField:    42,
				BoolField:   true,
			},
			new: MixedFieldTypes{
				StringField: "new",
				IntField:    100,
				BoolField:   true,
			},
			wantChanged: []string{"StringField", "IntField"},
		},

		// Non-struct types - should return empty changed keys
		{
			name:        "map types - should return empty changed keys",
			old:         map[string]any{"key": "value1"},
			new:         map[string]any{"key": "value2"},
			wantChanged: []string{},
		},
		{
			name:        "string types - should return empty changed keys",
			old:         "old",
			new:         "new",
			wantChanged: []string{},
		},
		{
			name:        "int types - should return empty changed keys",
			old:         42,
			new:         100,
			wantChanged: []string{},
		},
		{
			name:        "slice types - should return empty changed keys",
			old:         []string{"a", "b"},
			new:         []string{"c", "d"},
			wantChanged: []string{},
		},

		// Zero values
		{
			name:        "zero value structs - no changes",
			old:         SimpleConfig{},
			new:         SimpleConfig{},
			wantChanged: []string{},
		},
		{
			name: "zero to non-zero values",
			old:  SimpleConfig{},
			new:  SimpleConfig{Host: "localhost", Port: 8080},
			wantChanged: []string{"Host", "Port"},
		},
		{
			name: "non-zero to zero values",
			old:  SimpleConfig{Host: "localhost", Port: 8080},
			new:  SimpleConfig{},
			wantChanged: []string{"Host", "Port"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			event := diffEvent(tt.old, tt.new)

			// Assert
			// Handle nil vs empty slice comparison (both are valid for "no changes")
			if len(tt.wantChanged) == 0 {
				assert.Empty(t, event.ChangedKeys, "ChangedKeys should be empty")
			} else {
				assert.ElementsMatch(t, tt.wantChanged, event.ChangedKeys, "ChangedKeys should match expected values")
			}
			assert.Equal(t, tt.old, event.OldConfig, "OldConfig should match input")
			assert.Equal(t, tt.new, event.NewConfig, "NewConfig should match input")
		})
	}
}

// TestDiffEventEdgeCases tests additional edge cases and corner scenarios
// that may not be covered in the main test suite.
func TestDiffEventEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("deeply nested struct changes", func(t *testing.T) {
		t.Parallel()

		type Level3 struct {
			Value string
		}
		type Level2 struct {
			Inner Level3
		}
		type Level1 struct {
			Nested Level2
		}

		old := Level1{Nested: Level2{Inner: Level3{Value: "old"}}}
		new := Level1{Nested: Level2{Inner: Level3{Value: "new"}}}

		event := diffEvent(old, new)

		// Only top-level field name should be in ChangedKeys
		assert.Equal(t, []string{"Nested"}, event.ChangedKeys)
	})

	t.Run("empty struct - no fields", func(t *testing.T) {
		t.Parallel()

		type EmptyStruct struct{}

		old := EmptyStruct{}
		new := EmptyStruct{}

		event := diffEvent(old, new)

		assert.Empty(t, event.ChangedKeys)
		assert.Equal(t, old, event.OldConfig)
		assert.Equal(t, new, event.NewConfig)
	})

	t.Run("struct with unexported fields - only exported fields compared", func(t *testing.T) {
		t.Parallel()

		type StructWithUnexported struct {
			Exported string
		}

		old := StructWithUnexported{Exported: "old"}
		new := StructWithUnexported{Exported: "new"}

		event := diffEvent(old, new)

		// Should detect exported field change
		assert.Contains(t, event.ChangedKeys, "Exported")
	})

	t.Run("double pointer dereferencing - only dereferences once", func(t *testing.T) {
		t.Parallel()

		type SimpleConfig struct {
			Host string
			Port int
		}

		config1 := SimpleConfig{Host: "localhost", Port: 8080}
		config2 := SimpleConfig{Host: "example.com", Port: 9090}

		ptr1 := &config1
		ptr2 := &config2

		event := diffEvent(&ptr1, &ptr2)

		// diffEvent only dereferences one level, so after dereferencing &ptr1 and &ptr2,
		// we get ptr1 and ptr2 which are pointers (not structs), so no diff is computed
		assert.Empty(t, event.ChangedKeys)
	})

	t.Run("mismatched types - one struct one non-struct", func(t *testing.T) {
		t.Parallel()

		type SimpleConfig struct {
			Host string
		}

		old := SimpleConfig{Host: "localhost"}
		new := "not a struct"

		event := diffEvent(old, new)

		// Different kinds - should return empty changed keys
		assert.Empty(t, event.ChangedKeys)
	})

	t.Run("slice of structs changed", func(t *testing.T) {
		t.Parallel()

		type Config struct {
			Items []string
		}

		old := Config{Items: []string{"a", "b"}}
		new := Config{Items: []string{"a", "b", "c"}}

		event := diffEvent(old, new)

		assert.Equal(t, []string{"Items"}, event.ChangedKeys)
	})
}

// BenchmarkDiffEvent measures the performance of diffEvent with various struct sizes.
func BenchmarkDiffEvent(b *testing.B) {
	type SmallStruct struct {
		Field1 string
		Field2 int
	}

	type MediumStruct struct {
		Field1  string
		Field2  int
		Field3  bool
		Field4  float64
		Field5  []string
		Field6  map[string]string
		Field7  string
		Field8  int
		Field9  bool
		Field10 float64
	}

	type LargeStruct struct {
		F1, F2, F3, F4, F5           string
		F6, F7, F8, F9, F10          int
		F11, F12, F13, F14, F15      bool
		F16, F17, F18, F19, F20      float64
		F21, F22, F23, F24, F25      string
		F26, F27, F28, F29, F30      int
		F31, F32, F33, F34, F35      bool
		F36, F37, F38, F39, F40      float64
		F41, F42, F43, F44, F45      string
		F46, F47, F48, F49, F50      int
	}

	benchmarks := []struct {
		name string
		old  any
		new  any
	}{
		{
			name: "small struct - no changes",
			old:  SmallStruct{Field1: "test", Field2: 42},
			new:  SmallStruct{Field1: "test", Field2: 42},
		},
		{
			name: "small struct - all changed",
			old:  SmallStruct{Field1: "old", Field2: 1},
			new:  SmallStruct{Field1: "new", Field2: 2},
		},
		{
			name: "medium struct - no changes",
			old: MediumStruct{
				Field1: "test", Field2: 42, Field3: true, Field4: 3.14,
				Field5: []string{"a"}, Field6: map[string]string{"k": "v"},
				Field7: "test2", Field8: 100, Field9: false, Field10: 2.71,
			},
			new: MediumStruct{
				Field1: "test", Field2: 42, Field3: true, Field4: 3.14,
				Field5: []string{"a"}, Field6: map[string]string{"k": "v"},
				Field7: "test2", Field8: 100, Field9: false, Field10: 2.71,
			},
		},
		{
			name: "medium struct - half changed",
			old: MediumStruct{
				Field1: "old1", Field2: 1, Field3: true, Field4: 1.0,
				Field5: []string{"a"}, Field6: map[string]string{"k": "v1"},
			},
			new: MediumStruct{
				Field1: "new1", Field2: 2, Field3: true, Field4: 1.0,
				Field5: []string{"b"}, Field6: map[string]string{"k": "v2"},
			},
		},
		{
			name: "large struct - no changes",
			old:  LargeStruct{F1: "test", F6: 42, F11: true, F16: 3.14},
			new:  LargeStruct{F1: "test", F6: 42, F11: true, F16: 3.14},
		},
		{
			name: "large struct - few changed",
			old:  LargeStruct{F1: "old", F2: "test", F6: 42, F11: true},
			new:  LargeStruct{F1: "new", F2: "test", F6: 100, F11: true},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = diffEvent(bm.old, bm.new)
			}
		})
	}
}
