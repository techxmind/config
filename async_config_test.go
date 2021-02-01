package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAsyncConfig(t *testing.T) {
	tb := time.Now()
	tm := tb
	originFun := _now
	defer func() {
		_now = originFun
	}()
	_now = func() time.Time {
		return tm
	}

	ast := assert.New(t)

	asyncer := NewMockAsyncer(false)

	asyncKey := "async_key1"
	asyncer.Set(asyncKey, []byte(`{"custom":"custom"}`))
	cfg := NewAsyncConfig(asyncer, asyncKey, 1000*time.Millisecond, false)
	ast.Equal("custom", cfg.Get("custom"))
	// get count = 1
	ast.EqualValues(1, cfg.Get("ct"))
	ast.EqualValues(1, cfg.Get("ct"))

	//time.Sleep(1001 * time.Millisecond)
	tm = tb.Add(1001 * time.Millisecond)
	// get count = 2
	ast.EqualValues(2, cfg.Get("ct"))

	// 异步刷新
	tm = tb
	cfg2 := NewAsyncConfig(asyncer, asyncKey, 1000*time.Millisecond, true)
	// ct++ cause cfg2 initialition
	ast.EqualValues(3, cfg2.Get("ct"))
	//time.Sleep(1001 * time.Millisecond)
	tm = tb.Add(1001 * time.Millisecond)
	// expired, trigger refresh async
	ast.EqualValues(3, cfg2.Get("ct"))
	// wait async refresh completed
	time.Sleep(1 * time.Millisecond)
	ast.EqualValues(4, cfg2.Get("ct"))

	err := cfg2.Set("custom", "custom2")
	ast.Nil(err)
	ast.Equal("custom2", cfg2.Get("custom"))

	// test yaml & update notify
	asyncer = NewMockAsyncer(true)
	asyncKey = "async_key.yml"
	ast.Equal(T_YAML, asyncer.ContentType(asyncKey))
	asyncer.Set(asyncKey, []byte(`
a: 1
b:
  c: 2
  d: [3, 4]
`))
	cfg3 := NewAsyncConfig(asyncer, asyncKey, 1000*time.Millisecond, false)
	ast.EqualValues(1, cfg3.Get("a"))
	asyncer.Set(asyncKey, []byte(`
a: 2
`))
	time.Sleep(1 * time.Millisecond) // wait for update
	ast.EqualValues(2, cfg3.Get("a"))
}
