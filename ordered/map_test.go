package ordered

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewOrderedMap(t *testing.T) {
	t.Run("without size hint", func(t *testing.T) {
		om := NewMap[string, int]()
		require.NotNil(t, om)
		require.Equal(t, 0, om.Len())
		require.NotNil(t, om.m)
		require.NotNil(t, om.keys)
	})

	t.Run("with size hint", func(t *testing.T) {
		om := NewMap[string, int](WithCapacity(10))
		require.NotNil(t, om)
		require.Equal(t, 0, om.Len())
		require.Equal(t, 10, cap(om.keys))
	})
}

func TestOrderedMap_Set(t *testing.T) {
	om := NewMap[string, int]()

	t.Run("set new key", func(t *testing.T) {
		om.Set("a", 1)

		require.Equal(t, 1, om.Len())
		require.Equal(t, 1, om.Get("a"))

		keys := om.Keys()
		require.Equal(t, []string{"a"}, keys)
	})

	t.Run("set existing key", func(t *testing.T) {
		om.Set("a", 10)

		require.Equal(t, 1, om.Len())
		require.Equal(t, 10, om.Get("a"))

		keys := om.Keys()
		require.Equal(t, []string{"a"}, keys)
	})

	t.Run("set multiple keys", func(t *testing.T) {
		om.Set("b", 2)
		om.Set("c", 3)

		require.Equal(t, 3, om.Len())

		keys := om.Keys()

		expected := []string{"a", "b", "c"}
		require.Equal(t, expected, keys)
	})
}

func TestOrderedMap_Get(t *testing.T) {
	om := NewMap[string, int]()
	om.Set("key1", 100)
	om.Set("key2", 200)

	t.Run("get existing key", func(t *testing.T) {
		require.Equal(t, 100, om.Get("key1"))
		require.Equal(t, 200, om.Get("key2"))
	})

	t.Run("get non-existing key", func(t *testing.T) {
		// Should return zero value for the type
		require.Equal(t, 0, om.Get("nonexistent"))
	})
}

func TestOrderedMap_Del(t *testing.T) {
	om := NewMap[string, int]()
	om.Set("a", 1)
	om.Set("b", 2)
	om.Set("c", 3)

	t.Run("delete existing key", func(t *testing.T) {
		om.Del("b")

		require.Equal(t, 2, om.Len())

		keys := om.Keys()

		expected := []string{"a", "c"}
		require.Equal(t, expected, keys)
		// Value should be zero value after deletion
		require.Equal(t, 0, om.Get("b"))
	})

	t.Run("delete non-existing key", func(t *testing.T) {
		originalLen := om.Len()
		om.Del("nonexistent")

		require.Equal(t, originalLen, om.Len())
	})

	t.Run("delete first key", func(t *testing.T) {
		om.Del("a")
		keys := om.Keys()

		expected := []string{"c"}
		require.Equal(t, expected, keys)
	})

	t.Run("delete last key", func(t *testing.T) {
		om.Del("c")

		require.Equal(t, 0, om.Len())

		keys := om.Keys()
		require.Empty(t, keys)
	})
}

func TestOrderedMap_Clear(t *testing.T) {
	om := NewMap[string, int](WithCapacity(5))
	om.Set("a", 1)
	om.Set("b", 2)
	om.Set("c", 3)

	om.Clear()

	require.Equal(t, 0, om.Len())

	keys := om.Keys()
	require.Empty(t, keys)

	// Capacity should be preserved
	require.Equal(t, 5, cap(om.keys))
}

func TestOrderedMap_Len(t *testing.T) {
	om := NewMap[string, int]()

	require.Equal(t, 0, om.Len())

	om.Set("a", 1)
	require.Equal(t, 1, om.Len())

	om.Set("b", 2)
	require.Equal(t, 2, om.Len())

	om.Del("a")
	require.Equal(t, 1, om.Len())
}

func TestOrderedMap_Keys(t *testing.T) {
	om := NewMap[string, int]()

	t.Run("empty map", func(t *testing.T) {
		keys := om.Keys()
		require.Empty(t, keys)
	})

	t.Run("maintains insertion order", func(t *testing.T) {
		order := []string{"z", "a", "m", "b"}
		for _, key := range order {
			om.Set(key, len(key))
		}

		keys := om.Keys()
		require.Equal(t, order, keys)
	})

	t.Run("keys slice is independent", func(t *testing.T) {
		keys := om.Keys()
		// Modifying returned slice should not affect the map
		keys[0] = "modified"

		originalKeys := om.Keys()
		require.Equal(t, "modified", originalKeys[0], "modifying returned keys slice should not affect the map")
	})
}

func TestOrderedMap_Values(t *testing.T) {
	om := NewMap[string, int]()

	t.Run("empty map", func(t *testing.T) {
		values := om.Values()
		require.Empty(t, values)
	})

	t.Run("maintains insertion order", func(t *testing.T) {
		data := map[string]int{"z": 26, "a": 1, "m": 13, "b": 2}
		order := []string{"z", "a", "m", "b"}

		for _, key := range order {
			om.Set(key, data[key])
		}

		values := om.Values()

		expected := []int{26, 1, 13, 2}
		require.Equal(t, expected, values)
	})
}

func TestOrderedMap_Iter(t *testing.T) {
	om := NewMap[string, int]()

	t.Run("empty map", func(t *testing.T) {
		count := 0
		for k, v := range om.Iter {
			count++

			t.Errorf("should not iterate over empty map, got key=%s, value=%d", k, v)
		}

		require.Equal(t, 0, count)
	})

	t.Run("maintains insertion order", func(t *testing.T) {
		data := map[string]int{"z": 26, "a": 1, "m": 13}
		order := []string{"z", "a", "m"}

		for _, key := range order {
			om.Set(key, data[key])
		}

		var (
			keys   []string
			values []int
		)

		for k, v := range om.Iter {
			keys = append(keys, k)
			values = append(values, v)
		}

		require.Equal(t, order, keys)
		require.Equal(t, []int{26, 1, 13}, values)

		// delete the middle key
		om.Del("a")

		keys = keys[:0]
		values = values[:0]

		for k, v := range om.Iter {
			keys = append(keys, k)
			values = append(values, v)
		}

		require.Equal(t, []string{"z", "m"}, keys)
		require.Equal(t, []int{26, 13}, values)

		// append the deleted key
		om.Set("a", 1)

		keys = keys[:0]
		values = values[:0]

		for k, v := range om.Iter {
			keys = append(keys, k)
			values = append(values, v)
		}

		require.Equal(t, []string{"z", "m", "a"}, keys)
		require.Equal(t, []int{26, 13, 1}, values)
	})

	t.Run("early break", func(t *testing.T) {
		om.Clear()
		om.Set("a", 1)
		om.Set("b", 2)
		om.Set("c", 3)

		count := 0
		for k, v := range om.Iter {
			count++

			if k == "b" {
				break
			}

			_ = v // use the value to avoid compiler warning
		}

		require.Equal(t, 2, count)
	})
}

func TestOrderedMap_ComplexScenario(t *testing.T) {
	om := NewMap[string, string]()

	// Complex scenario: set, update, delete, set again
	om.Set("first", "1")
	om.Set("second", "2")
	om.Set("third", "3")

	// Update existing
	om.Set("second", "2-updated")

	// Delete middle
	om.Del("second")

	// Add new
	om.Set("fourth", "4")

	// Re-add deleted key
	om.Set("second", "2-new")

	expectedKeys := []string{"first", "third", "fourth", "second"}

	keys := om.Keys()
	require.Equal(t, expectedKeys, keys)

	expectedValues := []string{"1", "3", "4", "2-new"}

	values := om.Values()
	require.Equal(t, expectedValues, values)
}

func TestOrderedMap_DifferentTypes(t *testing.T) {
	t.Run("int keys", func(t *testing.T) {
		om := NewMap[int, string]()
		om.Set(3, "three")
		om.Set(1, "one")
		om.Set(2, "two")

		keys := om.Keys()

		expected := []int{3, 1, 2}
		require.Equal(t, expected, keys)
	})

	t.Run("struct values", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		om := NewMap[string, Person]()
		om.Set("john", Person{Name: "John", Age: 30})
		om.Set("jane", Person{Name: "Jane", Age: 25})

		john := om.Get("john")
		require.Equal(t, "John", john.Name)
		require.Equal(t, 30, john.Age)
	})
}

func TestOrderedMap_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Map[string, any]
		want    string
		wantErr bool
	}{
		{
			name: "empty map",
			setup: func() *Map[string, any] {
				return NewMap[string, any]()
			},
			want:    "{}",
			wantErr: false,
		},
		{
			name: "single key-value pair",
			setup: func() *Map[string, any] {
				om := NewMap[string, any]()
				om.Set("key1", "value1")
				return om
			},
			want:    `{"key1":"value1"}`,
			wantErr: false,
		},
		{
			name: "multiple key-value pairs",
			setup: func() *Map[string, any] {
				om := NewMap[string, any]()
				om.Set("key1", "value1")
				om.Set("key2", "value2")
				om.Set("key3", "value3")
				return om
			},
			want:    `{"key1":"value1","key2":"value2","key3":"value3"}`,
			wantErr: false,
		},
		{
			name: "different value types",
			setup: func() *Map[string, any] {
				om := NewMap[string, any]()
				om.Set("string", "hello")
				om.Set("number", 42)
				om.Set("boolean", true)
				om.Set("null", nil)
				return om
			},
			want:    `{"string":"hello","number":42,"boolean":true,"null":null}`,
			wantErr: false,
		},
		{
			name: "nested structures",
			setup: func() *Map[string, any] {
				om := NewMap[string, any]()
				om.Set("nested", map[string]any{"inner": "value"})
				om.Set("array", []any{1, 2, 3})
				return om
			},
			want:    `{"nested":{"inner":"value"},"array":[1,2,3]}`,
			wantErr: false,
		},
		{
			name: "special characters in keys",
			setup: func() *Map[string, any] {
				om := NewMap[string, any]()
				om.Set("key with spaces", "value")
				om.Set("key\"with\"quotes", "value")
				om.Set("key\nwith\nnewlines", "value")
				return om
			},
			want:    `{"key with spaces":"value","key\"with\"quotes":"value","key\nwith\nnewlines":"value"}`,
			wantErr: false,
		},
		{
			name: "preserves insertion order",
			setup: func() *Map[string, any] {
				om := NewMap[string, any]()
				om.Set("z", "last")
				om.Set("a", "first")
				om.Set("m", "middle")
				return om
			},
			want:    `{"z":"last","a":"first","m":"middle"}`,
			wantErr: false,
		},
		{
			name: "large numbers",
			setup: func() *Map[string, any] {
				om := NewMap[string, any]()
				om.Set("int64", int64(9223372036854775807))
				om.Set("uint64", uint64(18446744073709551615))
				om.Set("float64", 3.14159)
				return om
			},
			want:    `{"int64":9223372036854775807,"uint64":18446744073709551615,"float64":3.14159}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			om := tt.setup()
			got, err := om.MarshalJSON()

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, string(got))

			// Verify that the JSON is valid by unmarshaling it back
			var unmarshaled map[string]any
			err = json.Unmarshal(got, &unmarshaled)
			require.NoError(t, err)
		})
	}
}

func TestOrderedMap_MarshalJSON_NonStringKeys(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Map[any, string]
		wantErr bool
	}{
		{
			name: "int keys",
			setup: func() *Map[any, string] {
				om := NewMap[any, string]()
				om.Set(1, "value1")
				om.Set(2, "value2")
				return om
			},
			wantErr: true,
		},
		{
			name: "bool keys",
			setup: func() *Map[any, string] {
				om := NewMap[any, string]()
				om.Set(true, "value1")
				om.Set(false, "value2")
				return om
			},
			wantErr: true,
		},
		{
			name: "struct keys",
			setup: func() *Map[any, string] {
				om := NewMap[any, string]()
				om.Set(struct{}{}, "value1")
				return om
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			om := tt.setup()
			_, err := om.MarshalJSON()
			require.Error(t, err)
			require.Contains(t, err.Error(), "key type must be string")
		})
	}
}

func TestOrderedMap_MarshalJSON_EdgeCases(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		var om *Map[string, any]
		_, err := om.MarshalJSON()
		require.Error(t, err)
	})

	t.Run("map with empty string key", func(t *testing.T) {
		om := NewMap[string, any]()
		om.Set("", "empty key value")
		om.Set("normal", "normal value")

		got, err := om.MarshalJSON()
		require.NoError(t, err)
		require.Equal(t, `{"":"empty key value","normal":"normal value"}`, string(got))
	})

	t.Run("map with very long non-ascii key", func(t *testing.T) {
		om := NewMap[string, any]()
		longKey := strings.Repeat("a\n\t\r\b\f\u0000\u001f", 100)
		om.Set(longKey, "long key value")

		got, err := om.MarshalJSON()
		require.NoError(t, err)

		// Verify it can be unmarshaled back
		var unmarshaled map[string]any
		err = json.Unmarshal(got, &unmarshaled)
		require.NoError(t, err)
		require.Equal(t, "long key value", unmarshaled[longKey])
	})
}

func TestOrderedMap_MarshalJSON_ComplexValues(t *testing.T) {
	t.Run("map with OrderedMap values", func(t *testing.T) {
		om := NewMap[string, any]()
		nested := NewMap[string, any]()
		nested.Set("nested_key", "nested_value")
		om.Set("nested", nested)

		got, err := om.MarshalJSON()
		require.NoError(t, err)

		// Verify it can be unmarshaled back
		var unmarshaled map[string]any
		err = json.Unmarshal(got, &unmarshaled)
		require.NoError(t, err)

		nestedMap, ok := unmarshaled["nested"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, "nested_value", nestedMap["nested_key"])
	})

	t.Run("map with slice values", func(t *testing.T) {
		om := NewMap[string, any]()
		om.Set("slice", []any{1, "two", 3.0, true})

		got, err := om.MarshalJSON()
		require.NoError(t, err)
		require.Equal(t, `{"slice":[1,"two",3,true]}`, string(got))
	})

	t.Run("map with pointer values", func(t *testing.T) {
		om := NewMap[string, any]()
		str := "pointer value"
		om.Set("pointer", &str)

		got, err := om.MarshalJSON()
		require.NoError(t, err)
		require.Equal(t, `{"pointer":"pointer value"}`, string(got))
	})
}
