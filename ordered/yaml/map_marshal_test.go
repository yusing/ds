package ordered

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	baseom "github.com/yusing/ds/ordered"
)

func TestMarshalYAML_KeyTypeNotString(t *testing.T) {
	m := NewMap[int, int]()
	_, err := m.MarshalYAML()
	require.ErrorIs(t, err, baseom.ErrKeyTypeNotString)
}

func TestMarshalYAML_NilReceiver(t *testing.T) {
	var m *Map[string, any]
	_, err := m.MarshalYAML()
	require.ErrorIs(t, err, baseom.ErrNilOrderedMap)
}

func TestMarshalYAML_Empty(t *testing.T) {
	m := NewMap[string, any]()
	out, err := m.MarshalYAML()
	require.NoError(t, err)
	require.Equal(t, "{}", string(out))
}

func TestMarshalYAML_SimpleScalars(t *testing.T) {
	m := NewMap[string, any]()
	m.Set("a", 123)
	m.Set("b", "xyz")

	out, err := m.MarshalYAML()
	require.NoError(t, err)

	s := string(out)
	// Expect two lines, in insertion order
	lines := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	require.Len(t, lines, 2)
	require.Equal(t, "'a': 123", lines[0])
	// string marshaling may omit quotes for simple scalars; accept either quoted or unquoted
	require.True(t, lines[1] == "'b': xyz" || lines[1] == "'b': 'xyz'", "got line: %q", lines[1])
}

func TestMarshalYAML_QuotesInKeyAreEscaped(t *testing.T) {
	m := NewMap[string, any]()
	m.Set("a'b", 1)

	out, err := m.MarshalYAML()
	require.NoError(t, err)
	require.Equal(t, "'a''b': 1\n", string(out))
}

func TestMarshalYAML_MultilineValueIndented(t *testing.T) {
	m := NewMap[string, any]()
	// nested map to force multiline YAML
	m.Set("root", map[string]any{"x": 1, "y": 2})

	out, err := m.MarshalYAML()
	require.NoError(t, err)

	s := string(out)
	require.True(t, strings.HasPrefix(s, "'root':\n"))

	// All subsequent non-empty lines should be indented by yaml.DefaultIndentSpaces spaces
	lines := strings.Split(s, "\n")
	require.GreaterOrEqual(t, len(lines), 2)
	indentStr := strings.Repeat(" ", 2) // yaml.DefaultIndentSpaces is 2 in goccy/go-yaml
	for i := 1; i < len(lines); i++ {
		if lines[i] == "" { // last split or empty line
			continue
		}
		require.True(t, strings.HasPrefix(lines[i], indentStr), "line not indented: %q", lines[i])
	}
}
