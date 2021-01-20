package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/techxmind/go-utils/fileutil"
)

/*
test config map:
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
func TestConfig(t *testing.T) {
	ast := assert.New(t)
	cfg := newConfig()
	//ast.Implements((*Config)(nil), cfg)
	ast.Implements((*Configer)(nil), cfg)
	layer1 := NewMapConfig(getTestConfigMap())
	layer2 := NewMapConfig(getTestConfigMap())
	layer2.Set("layer_name", "layer2")
	cfg.AddLayer(DefaultLayerName, layer1)
	cfg.AddLayer("layer2", layer2)
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

	ast.Equal("", cfg.String("layer_name"))
	ast.Equal("layer2", cfg.Layer("layer2").String("layer_name"))
	ast.Equal(int64(0), cfg.Layer("layer2").Int("l2"))                   //l2在layer1中设置，在layer2中未定义
	ast.Equal(int64(8), cfg.Layer("layer2", DefaultLayerName).Int("l2")) //l2在layer1中设置，在layer2中未定义

	proxy := cfg.Layer("layer2")
	ast.Equal("layer2", proxy.String("layer_name"))

	m := cfg.Map("l1")
	ast.Equal("l61_value", m.String("l6.l61"))

	m = cfg.Map("not_exist")
	ast.Nil(m)

	s := new(struct {
		L15 int `json:"l15"`
	})
	err = cfg.Remarshal("l1", s)
	ast.Nil(err)
	ast.Equal(0, s.L15)
	err = cfg.Remarshal("not_exist", s)
	ast.NotNil(err)
}

func TestAddDefaultLayerName(t *testing.T) {
	t.Skip("Skipping test because of ZK dependency")
	ast := assert.New(t)
	layerName := "TestAddDefaultLayerName"
	key := "TestAddDefaultLayerName"
	layer := NewMapConfig(map[string]interface{}{
		key: true,
	})
	cfg := newConfig()
	cfg.AddLayer(layerName, layer)
	ast.False(cfg.Bool(key))
	cfg.AddDefaultLayerName(layerName)
	ast.True(cfg.Bool(key))
	cfg.RemoveDefaultLayerName(layerName)
	ast.False(cfg.Bool(key))
}

func TestStaticConfFile(t *testing.T) {
	t.Skip("Skipping test because of ZK dependency")
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected ioutil.TempDir error: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	confContent := []byte(`
{
	"c1" : "value1",
	"c2" : 2,
	"c3" : true
}
	`)

	confFile := filepath.Join(tmpdir, "go_test_conf.conf")
	err = ioutil.WriteFile(confFile, confContent, fileutil.PrivateFileMode)
	if err != nil {
		t.Fatalf("write conf file error: %v", err)
	}
	defer os.Remove(confFile)

	ast := assert.New(t)
	Merge(getTestConfigMap())
	ast.Equal("", String("c1"))
	initWithFile(confFile, false, 10*time.Millisecond, false)
	ast.Equal(1, len(_cfg.defaultLayerNames.Load().([]string)))
	ast.Equal("value1", String("c1"))
	ast.Equal(int64(2), Int("c2"))
	ast.Equal(true, Bool("c3"))
	ast.Equal(true, Bool("test_conf")) // from test config map
}

func TestDynamicConfFile(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected ioutil.TempDir error: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	confContent := []byte(`
{
	"c11" : "value1",
	"c22" : 2,
	"c33" : true
}
	`)

	confFile := filepath.Join(tmpdir, "go_test_conf2.conf")
	err = ioutil.WriteFile(confFile, confContent, fileutil.PrivateFileMode)
	if err != nil {
		t.Fatalf("write conf file error: %v", err)
	}
	defer os.Remove(confFile)
	defer _cfg.RemoveLayer("default-conf-file")

	ast := assert.New(t)
	Merge(getTestConfigMap())
	initWithFile(confFile, true, 10*time.Millisecond, false)
	//time.Sleep(11 * time.Millisecond)
	ast.Equal(2, len(_cfg.defaultLayerNames.Load().([]string)))
	ast.Equal("value1", String("c11"))
	ast.Equal(int64(2), Int("c22"))
	ast.Equal(true, Bool("c33"))
	ast.Equal(true, Bool("test_conf")) // from test config map

	confContent = []byte(`
{
	"c11" : "value1.1",
	"c22" : 2, // this is comment
	"c33" : true
}
	`)
	time.Sleep(1 * time.Second)
	err = ioutil.WriteFile(confFile, confContent, fileutil.PrivateFileMode)
	if err != nil {
		t.Fatalf("write conf file error: %v", err)
	}
	for i := 0; i <= 100; i++ { // wait for synchronization to complete
		time.Sleep(10 * time.Millisecond)
		if String("c11") == "value1.1" {
			break
		}
	}
	ast.Equal("value1.1", String("c11"))
}
