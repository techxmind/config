package config

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testAsyncer struct {
	sync.Mutex
	ct   int32
	data string
}

func (a *testAsyncer) Get(keyPath string) []byte {
	a.Lock()
	defer a.Unlock()
	var data interface{}
	json.Unmarshal([]byte(a.data), &data)
	m := data.(map[string]interface{})
	a.ct++
	setMapValue(m, "key", keyPath)
	setMapValue(m, "ct", a.ct)
	ret, _ := json.Marshal(m)
	logger.Infof("get async config[%s]:%s", keyPath, ret)
	return ret
}

func (a *testAsyncer) Set(keyPath string, value []byte) error {
	a.Lock()
	defer a.Unlock()
	a.data = string(value)
	logger.Infof("set async config[%s]:%s", keyPath, value)
	return nil
}

func (a *testAsyncer) Watch(keyPath string) chan bool {
	return nil
}

func TestAsyncConfig(t *testing.T) {
	ast := assert.New(t)

	asyncer := &testAsyncer{
		data: `{"custom":"custom"}`,
	}

	asyncKey := "async_key1"
	cfg := NewAsyncConfig(asyncer, asyncKey, 1000*time.Millisecond, false)
	ast.Equal("custom", cfg.Get("custom"))
	// get count = 1
	ast.EqualValues(1, cfg.Get("ct"))
	ast.EqualValues(1, cfg.Get("ct"))
	time.Sleep(1001 * time.Millisecond)
	// get count = 2
	ast.EqualValues(2, cfg.Get("ct"))

	// 异步刷新
	cfg2 := NewAsyncConfig(asyncer, asyncKey, 1000*time.Millisecond, true)
	// ct++ cause cfg2 initialition
	ast.EqualValues(3, cfg2.Get("ct"))
	time.Sleep(1001 * time.Millisecond)
	// expired, trigger refresh async
	ast.EqualValues(3, cfg2.Get("ct"))
	// wait async refresh completed
	time.Sleep(100 * time.Millisecond)
	ast.EqualValues(4, cfg2.Get("ct"))

	err := cfg2.Set("custom", "custom2")
	ast.Nil(err)
	ast.Equal("custom2", cfg2.Get("custom"))
}
