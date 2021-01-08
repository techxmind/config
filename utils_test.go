package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/techxmind/go-utils/object"
)

func TestJSONToMap(t *testing.T) {
	ast := assert.New(t)
	m, err := JSONToMap([]byte(`
{
"foo": "bar"
}
	`))

	ast.Nil(err)
	ast.NotNil(m)

	if m != nil {
		foo, ok := m["foo"]
		ast.True(ok)
		ast.Equal("bar", foo)
	}
}

func TestIsFalseStr(t *testing.T) {
	ast := assert.New(t)
	ast.True(isFalseStr(""))
	ast.True(isFalseStr("false"))
	ast.True(isFalseStr("0"))
	ast.False(isFalseStr("1"))
	ast.False(isFalseStr("true"))
	ast.False(isFalseStr("00"))
}

func TestSetMapValue(t *testing.T) {
	ast := assert.New(t)

	m := map[string]interface{}{
		"l1": map[string]interface{}{
			"l11": []interface{}{"a", "b", "c"},
			"l12": map[string]interface{}{
				"l13": "l13_value",
			},
		},
	}

	var err error
	var v interface{}
	err = setMapValue(m, "l1.l14", "l14_value")
	ast.Nil(err)
	v, _ = object.GetValue(m, "l1.l14")
	ast.Equal(v, "l14_value")

	err = setMapValue(m, "l1.l12", []interface{}{"a"})
	ast.Nil(err)
	v, _ = object.GetValue(m, "l1.l12")
	ast.IsType([]interface{}{}, v)

	err = setMapValue(m, "l1.l11.0", "d")
	ast.Nil(err)
	v, _ = object.GetValue(m, "l1.l11")
	ast.Equal([]interface{}{"d", "b", "c"}, v)

	err = setMapValue(m, "l1.l11.n", "a")
	ast.NotNil(err)

	err = setMapValue(m, "l1.l11.3", "a")
	ast.NotNil(err)
}

func TestMergeMap(t *testing.T) {
	ast := assert.New(t)

	m1 := map[string]interface{}{
		"a": map[string]interface{}{
			"a1": map[string]interface{}{
				"a11": "a11_value",
				"a12": "a12_value",
			},
			"a2": []interface{}{1, 2, 3},
		},
		"b": map[string]interface{}{
			"b1": "b1_value",
		},
		"c": 13,
	}

	m2 := map[string]interface{}{
		"a": map[string]interface{}{
			"a1": map[string]interface{}{
				"a11": "a11_value_2",
				"a13": "a13_value",
			},
			"a2": []interface{}{2, 2, 3},
		},
		"b": 13,
		"c": map[string]interface{}{
			"b1": "b1_value",
		},
	}

	expectMap := map[string]interface{}{
		"a": map[string]interface{}{
			"a1": map[string]interface{}{
				"a11": "a11_value_2",
				"a12": "a12_value",
				"a13": "a13_value",
			},
			"a2": []interface{}{2, 2, 3},
		},
		"b": 13,
		"c": map[string]interface{}{
			"b1": "b1_value",
		},
	}

	mergeMap(m1, m2)
	ast.Equal(expectMap, m1)
}
