package ordered

import (
	"bytes"
	"reflect"
	"strings"
	"unsafe"

	"github.com/goccy/go-yaml"
	"github.com/yusing/ds/ordered"
)

type Map[K comparable, V any] struct {
	// keep the anonymous field private
	omap[K, V]
}

type omap[K comparable, V any] = ordered.Map[K, V]

type Option = ordered.Option

func NewMap[K comparable, V any](opts ...Option) *Map[K, V] {
	return &Map[K, V]{omap: *ordered.NewMap[K, V](opts...)}
}

func (o *Map[K, V]) MarshalYAML() ([]byte, error) {
	if reflect.TypeFor[K]().Kind() != reflect.String {
		return nil, ordered.ErrKeyTypeNotString
	}

	if o == nil {
		return nil, ordered.ErrNilOrderedMap
	}

	if o.Len() == 0 {
		return []byte("{}"), nil
	}

	keys := o.Keys()

	// can just convert it directly to string slice to avoid unnecessary allocation
	strKeys := *(*[]string)(unsafe.Pointer(&keys))

	// handle root keys to preserve the insertion order
	buf := bytes.NewBuffer(make([]byte, 0, o.Len()*20))
	for i, keyStr := range strKeys {
		// write YAML key (quote conservatively to avoid special-char pitfalls)
		writeYAMLQuotedString(buf, keyStr)
		buf.WriteByte(':')

		// encode value using YAML so nested structs/slices are handled correctly
		v := o.Get(keys[i])
		vb, err := yaml.Marshal(v)
		if err != nil {
			return nil, err
		}

		// yaml.Marshal adds a trailing newline; trim for inline usage/indent logic
		vb = bytes.TrimSuffix(vb, []byte{'\n'})

		if bytes.Contains(vb, []byte{'\n'}) {
			// multiline value -> start on next line and indent each line
			buf.WriteByte('\n')
			for line := range bytes.Lines(vb) {
				buf.WriteString(indent)
				buf.Write(line)
				buf.WriteByte('\n')
			}
		} else {
			// single line -> keep inline
			if len(vb) > 0 {
				buf.WriteByte(' ')
				buf.Write(vb)
			}
			buf.WriteByte('\n')
		}
	}

	return buf.Bytes(), nil
}

var indent = strings.Repeat(" ", yaml.DefaultIndentSpaces)

func writeYAMLQuotedString(buf *bytes.Buffer, s string) {
	buf.WriteByte('\'')
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\'' {
			if i > start {
				buf.WriteString(s[start:i])
			}
			// escape single quote in YAML single-quoted scalar by doubling it
			buf.WriteString("''")
			start = i + 1
		}
	}
	if start < len(s) {
		buf.WriteString(s[start:])
	}
	buf.WriteByte('\'')
}
