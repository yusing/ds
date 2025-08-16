package ordered

import (
	"bytes"
	"encoding/json"
	"errors"
	"maps"
	"reflect"
	"slices"
	"unsafe"
)

type Map[K comparable, V any] struct {
	m    map[K]V
	keys []K
}

var (
	ErrKeyTypeNotString = errors.New("key type must be string")
	ErrNilOrderedMap    = errors.New("calling MarshalJSON on nil OrderedMap")
)

func NewMap[K comparable, V any](opts ...Option) *Map[K, V] {
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
	if reflect.TypeFor[K]().Kind() != reflect.String {
		return nil, ErrKeyTypeNotString
	}

	if o == nil {
		return nil, ErrNilOrderedMap
	}

	if o.Len() == 0 {
		return []byte("{}"), nil
	}

	// can just convert it directly to string slice to avoid unnecessary allocation
	strKeys := *(*[]string)(unsafe.Pointer(&o.keys))

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
