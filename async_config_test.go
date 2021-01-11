package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAsyncConfig(t *testing.T) {
	ast := assert.New(t)

	asyncer := NewMockAsyncer()

	asyncKey := "async_key1"
	asyncer.Set(asyncKey, []byte(`{"custom":"custom"}`))
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
