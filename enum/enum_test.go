package enum

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

type CustomEnum int

const (
	Unknown CustomEnum = 0
	Val1    CustomEnum = 1
	Val2    CustomEnum = 2
)

var customEnum = New().
	Add(Unknown, "unknown").
	Add(Val1, "val1").
	Add(Val2, "val2")

func TestStringKeys(t *testing.T) {

	require.Equal(t, []string{}, New().StringKeys())

	keys := customEnum.StringKeys()
	sort.Strings(keys)
	require.Equal(t, []string{"unknown", "val1", "val2"}, keys)
}

func TestGetByString(t *testing.T) {

	get := func(src string) CustomEnum {
		mode, ok := customEnum.GetByString(src)
		if !ok {
			return Unknown
		}
		return mode.(CustomEnum)
	}

	require.Equal(t, Unknown, get("-"))
	require.Equal(t, Unknown, get("unknown"))
	require.Equal(t, Val1, get("val1"))
	require.Equal(t, Val2, get("val2"))
}

func TestGetByIndex(t *testing.T) {

	get := func(src interface{}) string {
		str, ok := customEnum.GetByIndex(src)
		if !ok {
			return ""
		}
		return str
	}

	require.Equal(t, "", get(-1))
	require.Equal(t, "unknown", get(Unknown))
	require.Equal(t, "val1", get(Val1))
	require.Equal(t, "val2", get(Val2))
}
