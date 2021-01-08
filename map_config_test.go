package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTestConfigMap() map[string]interface{} {
	m, _ := JSONToMap([]byte(`{
"test_conf" : true,
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
}	`))
	return m
}

func TestMapConfig(t *testing.T) {
	cfg := NewMapConfig(getTestConfigMap())
	testMap(t, cfg)
}

func TestNoSyncMapConfig(t *testing.T) {
	cfg := NewMapConfig(getTestConfigMap(), false)
	testMap(t, cfg)
}

func testMap(t *testing.T, cfg *MapConfig) {
	ast := assert.New(t)

	ast.Equal("l112_value", cfg.String("l1.l11.l112"))
	ast.Equal(uint64(1), cfg.Uint("l1.l11.l111.l1111.0"))
	ast.Equal(int64(1), cfg.Int("l1.l11.l111.l1111.0"))
	ast.Equal("1", cfg.String("l1.l11.l111.l1111.0"))
	ast.InDelta(0.12, cfg.Float("l1.l12"), 0.000001)
	ast.Equal(false, cfg.Bool("l1.l13"))
	ast.Equal(false, cfg.Bool("l1.l15"))
	ast.Equal(true, cfg.Bool("l1.l14"))
	ast.Equal(int64(6), cfg.IntDefault("l2", 6))
	err := cfg.Set("l2", 7)
	ast.Nil(err)
	ast.Equal(int64(7), cfg.IntDefault("l2", 6))
	err = cfg.Set("l2", 8)
	ast.Nil(err)
	ast.Equal(int64(8), cfg.Int("l2"))
	err = cfg.Set(RootKey, map[string]interface{}{
		"l3": 3,
		"l4": "l4_value",
	})
	ast.Nil(err)
	ast.Equal(int64(3), cfg.Int("l3"))
	ast.Equal("l4_value", cfg.String("l4"))
	ast.Equal(int64(0), cfg.Int("l5")) // 不存在的配置，返回默认值
	err = cfg.Set("l1.l6.l61", "l61_value")
	ast.Nil(err)
	ast.Equal("l61_value", cfg.String("l1.l6.l61"))
	err = cfg.Set("l1.l14.l141", "l141_value")
	ast.NotNil(err) //l1.l14已存在，且结构冲突

	err = cfg.Merge(map[string]interface{}{
		"mg1": "mg1_value",
		"l1": map[string]interface{}{
			"l8": "l8_value",
		},
	})
	ast.Nil(err)
	ast.Equal("mg1_value", cfg.String("mg1"))
	ast.Equal("l8_value", cfg.String("l1.l8"))
	ast.Equal("l112_value", cfg.String("l1.l11.l112"))

	subCfg := cfg.Map("l1")
	ast.Equal("l112_value", subCfg.String("l11.l112"))

	s := new(struct {
		L112 string `json:"l112"`
	})

	err = cfg.Remarshal("l1.l11", s)
	ast.Nil(err)
	ast.Equal("l112_value", s.L112)
}

func BenchmarkMapConfigSimpleGet(b *testing.B) {
	cfg := NewMapConfig(getTestConfigMap())

	for i := 0; i < b.N; i++ {
		cfg.Int("l1.l14")
	}
}

func BenchmarkMapConfigDeepGet(b *testing.B) {
	cfg := NewMapConfig(getTestConfigMap())

	for i := 0; i < b.N; i++ {
		cfg.Int("l1.l11.l111.l1111.0")
	}
}

func BenchmarkMapConfigSet(b *testing.B) {
	ast := assert.New(b)

	cfg := NewMapConfig(getTestConfigMap())

	for i := 0; i < b.N; i++ {
		err := cfg.Set(fmt.Sprintf("n%d", i%200), "value")
		ast.Nil(err)
	}
}

func BenchmarkMapConfigNoSyncModeSet(b *testing.B) {
	ast := assert.New(b)

	cfg := NewMapConfig(getTestConfigMap(), false)

	for i := 0; i < b.N; i++ {
		err := cfg.Set(fmt.Sprintf("n%d", i%200), "value")
		ast.Nil(err)
	}
}
