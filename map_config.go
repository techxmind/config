package config

import (
	"sync"
	"sync/atomic"

	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"

	"github.com/techxmind/go-utils/object"
)

type MapConfig struct {
	ConfigHelper
}

// syncMode : 是否同步模式（并发场景），默认为同步模式
// 非同步模式Set性能会非常高，适用于一些在非并发场景的局部变量的使用
func NewMapConfig(m map[string]interface{}, syncMode ...bool) *MapConfig {
	if m == nil {
		m = make(map[string]interface{})
	}

	smode := true
	if len(syncMode) > 0 {
		smode = syncMode[0]
	}

	mc := &mapConfig{
		syncMode:  smode,
		notifiers: make([]chan struct{}, 0),
	}
	mc.m.Store(m)

	return &MapConfig{
		ConfigHelper: ConfigHelper{
			Configer: mc,
		},
	}
}

type mapConfig struct {
	sync.Mutex
	syncMode  bool
	notifiers []chan struct{}
	m         atomic.Value //map[string]interface{}
}

func (m *mapConfig) Get(keyPath string) interface{} {
	if keyPath == RootKey {
		return m.m.Load()
	}

	val, ok := object.GetValue(m.m.Load(), keyPath)
	if !ok {
		return nil
	}

	return val
}

// Set 设置配置
//
// 同步模式性能较低(359913 ns/op)：每次调会clone一个新的副本，并在副本上更新，替换原配置map
func (m *mapConfig) Set(keyPath string, value interface{}) error {

	var newMap map[string]interface{}

	if m.syncMode {
		m.Lock()
		defer m.Unlock()

		newMap = deepcopy.Copy(m.m.Load()).(map[string]interface{})
	} else {
		newMap = m.m.Load().(map[string]interface{})
	}

	if keyPath == RootKey {
		if vm, ok := value.(map[string]interface{}); ok {
			mergeMap(newMap, vm)
		} else {
			return errors.Errorf("merge map error: value is not map[string]interface{}:%v", value)
		}
	} else {
		if err := setMapValue(newMap, keyPath, value); err != nil {
			return errors.Wrapf(err, "set map error key=%s", keyPath)
		}
	}

	if m.syncMode {
		m.m.Store(newMap)
	}

	m.notify()

	return nil
}

func (m *mapConfig) notify() {
	for _, notifier := range m.notifiers {
		select {
		case notifier <- struct{}{}:
		default:
		}
	}
}

func (m *mapConfig) Watch(notifier chan struct{}) {
	m.Lock()
	defer m.Unlock()
	m.notifiers = append(m.notifiers, notifier)
}
