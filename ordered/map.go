package ordered

import (
	"bytes"
	"encoding/json"
	"errors"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"unsafe"
)

type MapKey interface {
	string | int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

type Map[K MapKey, V any] struct {
	m    map[K]V
	keys []K
}

var (
	ErrNilOrderedMap = errors.New("calling MarshalJSON on nil OrderedMap")
)

func NewMap[K MapKey, V any](opts ...Option) *Map[K, V] {
	var opt option
	for _, o := range opts {
		o(&opt)
	}
	m := &Map[K, V]{
		m:    make(map[K]V, opt.capacity),
		keys: make([]K, 0, opt.capacity),
	}
	return m
}

func (o *Map[K, V]) Set(key K, value V) {
	oldSize := len(o.m)
	o.m[key] = value
	if len(o.m) > oldSize { // new key added
		o.keys = append(o.keys, key)
	}
}

func (o *Map[K, V]) Get(key K) V {
	return o.m[key]
}

func (o *Map[K, V]) TryGet(key K) (V, bool) {
	value, ok := o.m[key]
	return value, ok
}

func (o *Map[K, V]) Contains(key K) bool {
	_, ok := o.m[key]
	return ok
}

func (o *Map[K, V]) Del(key K) {
	oldSize := o.Len()
	if oldSize == 0 {
		return
	}

	delete(o.m, key)
	if len(o.m) == oldSize { // key not found
		return
	}

	if oldSize == 1 {
		o.keys = o.keys[:0]
		return
	}

	idx := slices.Index(o.keys, key)
	if idx == -1 {
		panic("race condition in OrderedMap.Del")
	}

	o.keys = slices.Delete(o.keys, idx, idx+1)
}

func (o *Map[K, V]) Len() int {
	return len(o.keys)
}

func (o *Map[K, V]) Keys() []K {
	return o.keys
}

func (o *Map[K, V]) Values() []V {
	values := make([]V, len(o.keys))
	for i, key := range o.keys {
		values[i] = o.m[key]
	}
	return values
}

func (o *Map[K, V]) Iter(yield func(key K, value V) bool) {
	for _, key := range o.keys {
		if !yield(key, o.m[key]) {
			break
		}
	}
}

func (o *Map[K, V]) IterKeys(yield func(key K) bool) {
	for _, key := range o.keys {
		if !yield(key) {
			break
		}
	}
}

func (o *Map[K, V]) IterValues(yield func(value V) bool) {
	for _, key := range o.keys {
		if !yield(o.m[key]) {
			break
		}
	}
}

func (o *Map[K, V]) Reverse() {
	slices.Reverse(o.keys)
}

func (o *Map[K, V]) Clear() {
	clear(o.m)
	o.keys = o.keys[:0]
}

func (o *Map[K, V]) Clone() *Map[K, V] {
	return &Map[K, V]{
		m:    maps.Clone(o.m),
		keys: slices.Clone(o.keys),
	}
}

func (o *Map[K, V]) MarshalJSON() ([]byte, error) {
	if o == nil {
		return nil, ErrNilOrderedMap
	}

	if o.Len() == 0 {
		return []byte("{}"), nil
	}

	// can just convert it directly to string slice to avoid unnecessary allocation
	var strKeys []string
	switch reflect.TypeFor[K]().Kind() {
	case reflect.String:
		strKeys = *(*[]string)(unsafe.Pointer(&o.keys))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		strKeys = make([]string, len(o.keys))
		for i, key := range o.keys {
			strKeys[i] = strconv.FormatInt(reflect.ValueOf(key).Int(), 10)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		strKeys = make([]string, len(o.keys))
		for i, key := range o.keys {
			strKeys[i] = strconv.FormatUint(reflect.ValueOf(key).Uint(), 10)
		}
	default: // float32 or float64
		strKeys = make([]string, len(o.keys))
		for i, key := range o.keys {
			strKeys[i] = strconv.FormatFloat(reflect.ValueOf(key).Float(), 'f', -1, 64)
		}
	}

	// handle root keys to preserve the insertion order
	buf := bytes.NewBuffer(make([]byte, 0, o.Len()*20))
	// using json.Encoder instead of json.Marshal
	// because we don't want to allocate a new byte slice for every key and value
	je := json.NewEncoder(buf)

	buf.WriteByte('{')
	for i, key := range strKeys {
		if i > 0 {
			buf.WriteByte(',')
		}

		writeEscapedString(buf, key)
		buf.WriteByte(':')

		err := je.Encode(o.m[o.keys[i]])
		if err != nil {
			return nil, err
		}

		buf.Truncate(buf.Len() - 1) // remove the trailing newline that Encode adds
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func MapEquals[K MapKey, V comparable](a, b *Map[K, V]) bool {
	return maps.Equal(a.m, b.m)
}

// OrderedMapMerge merges the given ordered maps into a new ordered map.
// Similar to array_merge in PHP.
func MapMerge[K MapKey, V any](m ...*Map[K, V]) *Map[K, V] {
	if len(m) == 0 {
		return NewMap[K, V]()
	}

	if len(m) == 1 {
		return m[0].Clone() // Clone to avoid modifying the original
	}

	var (
		keys   []K
		values map[K]V
	)

	totalSize := 0
	for _, om := range m {
		totalSize += om.Len()
	}

	seen := make(map[K]bool, totalSize)
	values = make(map[K]V, totalSize)
	keys = make([]K, 0, totalSize)

	for _, om := range m {
		for _, key := range om.keys {
			if !seen[key] {
				seen[key] = true
				keys = append(keys, key)
				values[key] = om.m[key]
			}
		}
	}

	return &Map[K, V]{
		m:    values,
		keys: keys,
	}
}
