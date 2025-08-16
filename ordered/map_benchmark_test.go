package ordered

import (
	"maps"
	"strconv"
	"testing"
)

func BenchmarkGet(b *testing.B) {
	const size = 1000

	b.Run("orderedmap", func(b *testing.B) {
		om := NewMap[string, int]()

		for i := range size {
			om.Set("key"+strconv.Itoa(i), i)
		}

		i := 0
		for b.Loop() {
			// 1000 exists and 500 not exists
			om.Get("key" + strconv.Itoa(i%(size*2)))
			i++
		}
	})

	b.Run("map", func(b *testing.B) {
		m := make(map[string]int, size)

		for i := range size {
			m["key"+strconv.Itoa(i)] = i
		}

		i := 0
		for b.Loop() {
			// 1000 exists and 500 not exists
			_ = m["key"+strconv.Itoa(i%(size*2))]
			i++
		}
	})
}

func BenchmarkSet(b *testing.B) {
	const size = 100
	b.Run("orderedmap", func(b *testing.B) {
		om := NewMap[string, int](WithCapacity(size))
		for b.Loop() {
			for i := range size {
				om.Set("key"+strconv.Itoa(i), i)
			}
			om.Clear()
		}
	})

	b.Run("map", func(b *testing.B) {
		m := make(map[string]int, size)
		for b.Loop() {
			for i := range size {
				m["key"+strconv.Itoa(i)] = i
			}
			clear(m)
		}
	})
}

func BenchmarkDel(b *testing.B) {
	const size = 1000

	b.Run("orderedmap", func(b *testing.B) {
		om := NewMap[string, int](WithCapacity(size))
		for i := range size {
			om.Set("key"+strconv.Itoa(i), i)
		}

		for b.Loop() {
			b.StopTimer()
			omClone := om.Clone()
			b.StartTimer()
			for i := range size {
				omClone.Del("key" + strconv.Itoa(i))
			}
		}
	})

	b.Run("map", func(b *testing.B) {
		m := make(map[string]int, size)
		for i := range size {
			m["key"+strconv.Itoa(i)] = i
		}

		for b.Loop() {
			b.StopTimer()
			mClone := maps.Clone(m)
			b.StartTimer()
			for i := range size {
				delete(mClone, "key"+strconv.Itoa(i))
			}
		}
	})
}

func BenchmarkIter(b *testing.B) {
	const size = 100000

	om := NewMap[string, int](WithCapacity(size))
	for i := range size {
		om.Set("key"+strconv.Itoa(i), i)
	}

	m := make(map[string]int, size)
	for i := range size {
		m["key"+strconv.Itoa(i)] = i
	}

	b.Run("orderedmap", func(b *testing.B) {
		for k, v := range om.Iter {
			_ = k + strconv.Itoa(v)
		}
	})

	b.Run("map", func(b *testing.B) {
		for k, v := range m {
			_ = k + strconv.Itoa(v)
		}
	})
}
