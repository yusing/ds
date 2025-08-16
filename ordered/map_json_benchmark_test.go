package ordered_test

import (
	"encoding/json"
	"strconv"
	"testing"

	. "github.com/yusing/ds/ordered"
)

type mapValue struct {
	Key   string
	Value int
}

func BenchmarkMarshalJSON(b *testing.B) {
	b.Run("orderedmap", func(b *testing.B) {
		om := NewMap[string, any](WithCapacity(1000))

		for i := range 1000 {
			om.Set("key"+strconv.Itoa(i), mapValue{
				Key:   "key" + strconv.Itoa(i),
				Value: i,
			})
		}

		b.ResetTimer()

		for b.Loop() {
			_, err := om.MarshalJSON()
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("map", func(b *testing.B) {
		m := make(map[string]any, 1000)

		for i := range 1000 {
			m["key"+strconv.Itoa(i)] = mapValue{
				Key:   "key" + strconv.Itoa(i),
				Value: i,
			}
		}

		b.ResetTimer()

		for b.Loop() {
			_, err := json.Marshal(m)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
