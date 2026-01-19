package ordered

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strconv"
	"unsafe"
)

type MapKey interface {
	~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

type Map[K MapKey, V any] struct {
	m    map[K]V
	keys []K
}

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

func (m *Map[K, V]) Set(key K, value V) {
	oldSize := len(m.m)
	m.m[key] = value
	if len(m.m) > oldSize { // new key added
		m.keys = append(m.keys, key)
	}
}

func (m *Map[K, V]) Get(key K) V {
	return m.m[key]
}

func (m *Map[K, V]) TryGet(key K) (V, bool) {
	value, ok := m.m[key]
	return value, ok
}

func (m *Map[K, V]) Contains(key K) bool {
	_, ok := m.m[key]
	return ok
}

func (m *Map[K, V]) Del(key K) {
	oldSize := m.Len()
	if oldSize == 0 {
		return
	}

	delete(m.m, key)
	if len(m.m) == oldSize { // key not found
		return
	}

	if oldSize == 1 {
		m.keys = m.keys[:0]
		return
	}

	idx := slices.Index(m.keys, key)
	if idx == -1 {
		panic("race condition in OrderedMap.Del")
	}

	m.keys = slices.Delete(m.keys, idx, idx+1)
}

func (m *Map[K, V]) Len() int {
	return len(m.keys)
}

func (m *Map[K, V]) Keys() []K {
	return m.keys
}

func (m *Map[K, V]) Values() []V {
	values := make([]V, len(m.keys))
	for i, key := range m.keys {
		values[i] = m.m[key]
	}
	return values
}

func (m *Map[K, V]) Iter(yield func(key K, value V) bool) {
	for _, key := range m.keys {
		if !yield(key, m.m[key]) {
			break
		}
	}
}

func (m *Map[K, V]) IterKeys(yield func(key K) bool) {
	for _, key := range m.keys {
		if !yield(key) {
			break
		}
	}
}

func (m *Map[K, V]) IterValues(yield func(value V) bool) {
	for _, key := range m.keys {
		if !yield(m.m[key]) {
			break
		}
	}
}

func (m *Map[K, V]) Reverse() {
	slices.Reverse(m.keys)
}

func (m *Map[K, V]) Clear() {
	clear(m.m)
	m.keys = m.keys[:0]
}

func (m *Map[K, V]) Clone() *Map[K, V] {
	return &Map[K, V]{
		m:    maps.Clone(m.m),
		keys: slices.Clone(m.keys),
	}
}

// MarshalJSON implements the json.Marshaler interface.
func (m *Map[K, V]) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}

	if m.Len() == 0 {
		return []byte("{}"), nil
	}

	// can just convert it directly to string slice to avoid unnecessary allocation
	var strKeys []string
	switch reflect.TypeFor[K]().Kind() {
	case reflect.String:
		strKeys = *(*[]string)(unsafe.Pointer(&m.keys))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		strKeys = make([]string, len(m.keys))
		for i, key := range m.keys {
			strKeys[i] = strconv.FormatInt(reflect.ValueOf(key).Int(), 10)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		strKeys = make([]string, len(m.keys))
		for i, key := range m.keys {
			strKeys[i] = strconv.FormatUint(reflect.ValueOf(key).Uint(), 10)
		}
	default: // float32 or float64
		strKeys = make([]string, len(m.keys))
		for i, key := range m.keys {
			strKeys[i] = strconv.FormatFloat(reflect.ValueOf(key).Float(), 'f', -1, 64)
		}
	}

	// handle root keys to preserve the insertion order
	buf := bytes.NewBuffer(make([]byte, 0, m.Len()*20))
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

		err := je.Encode(m.m[m.keys[i]])
		if err != nil {
			return nil, err
		}

		buf.Truncate(buf.Len() - 1) // remove the trailing newline that Encode adds
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// The key order follows the json order.
func (m *Map[K, V]) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))

	// expect the opening brace '{'
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expected JSON object opening '{', got %v", t)
	}

	var convertKey func(string) (K, error)
	switch reflect.TypeFor[K]().Kind() {
	case reflect.String:
		convertKey = func(key string) (K, error) {
			return *(*K)(unsafe.Pointer(&key)), nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		convertKey = func(key string) (K, error) {
			i, err := strconv.ParseInt(key, 10, 64)
			if err != nil {
				var zero K
				return zero, err
			}
			return *(*K)(unsafe.Pointer(&i)), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		convertKey = func(key string) (K, error) {
			i, err := strconv.ParseUint(key, 10, 64)
			if err != nil {
				var zero K
				return zero, err
			}
			return *(*K)(unsafe.Pointer(&i)), nil
		}
	default: // float32 or float64
		convertKey = func(key string) (K, error) {
			f, err := strconv.ParseFloat(key, 64)
			if err != nil {
				var zero K
				return zero, err
			}
			return *(*K)(unsafe.Pointer(&f)), nil
		}
	}

	// iterate over the tokens until we hit the closing brace
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return err
		}
		key, ok := t.(string)
		if !ok {
			return fmt.Errorf("expected string key, got %T", t)
		}

		// handle complex nested values (objects, arrays) automatically into type V.
		var value V
		if err := dec.Decode(&value); err != nil {
			return err
		}

		convertedKey, err := convertKey(key)
		if err != nil {
			return fmt.Errorf("unexpected key type: %w", err)
		}

		m.Set(convertedKey, value)
	}

	// consume the closing brace '}'
	t, err = dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '}' {
		return fmt.Errorf("expected JSON object closing '}', got %v", t)
	}

	return nil
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
