package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testL111 struct {
	L1111 []int `json:"l1111"`
}

/*
test config
{
"l1" : {
	"l11" : {
		"l111" : {
			"l1111" : [1, 3, 5]
		},
		"l112" : "l112_value"
	},
	"l12" : 0.12,
	"l13" : false,
	"l14" : 14,
	"l15" : 0
}
}
*/
func TestProxy(t *testing.T) {
	ast := assert.New(t)
	cfg := newConfig()
	cfg.AddLayer("custom", NewMapConfig(getTestConfigMap()))
	ast.Equal("", cfg.String("l1.l11.l112"))
	ast.Equal("l112_value", cfg.Layer("custom").String("l1.l11.l112"))
	proxy := cfg.Layer("custom")

	proxy.Set("l16", "l16_value")
	ast.Equal("l16_value", proxy.String("l16"))
	ast.Equal("l17_value", proxy.StringDefault("l17", "l17_value"))

	ast.Equal([]byte("l16_value"), proxy.Bytes("l16"))
	ast.Equal([]byte("l17_value_2"), proxy.BytesDefault("l17", []byte("l17_value_2")))

	ast.Equal(int64(14), proxy.Int("l1.l14"))
	ast.Equal(int64(4), proxy.IntDefault("l2", 4))

	ast.Equal(uint64(14), proxy.Uint("l1.l14"))
	ast.Equal(uint64(4), proxy.UintDefault("l2", 4))

	ast.InDelta(float64(0.12), proxy.Float("l1.l12"), .0000001)
	ast.InDelta(float64(0.13), proxy.FloatDefault("l2", 0.13), .0000001)

	ast.Equal(false, proxy.Bool("l1.l13"))
	ast.Equal(false, proxy.BoolDefault("l1.l13", true))
	ast.Equal(false, proxy.Bool("l1.l15"))
	ast.Equal(false, proxy.Bool("l1.l17"))
	ast.Equal(true, proxy.BoolDefault("l1.l17", true))
	ast.Equal(true, proxy.Bool("l1.l14"))

	m := proxy.Map("l1.l11")
	ast.Equal("l112_value", m.String("l112"))

	v := new(testL111)
	err := proxy.Remarshal("l1.l11.l111", v)
	ast.Nil(err)
	ast.NotNil(v)
	ast.Equal([]int{1, 3, 5}, v.L1111)

	proxy.Merge(map[string]interface{}{
		"l3": "l3_value",
		"l4": "l4_value",
	})
	ast.Equal("l3_value", proxy.String("l3"))
	ast.Equal("l4_value", proxy.String("l4"))
}
