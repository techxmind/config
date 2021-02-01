package config

import (
	"strings"
	"sync"
	"sync/atomic"
)

type MockAsyncer struct {
	ct            int32
	data          sync.Map
	notifyChans   sync.Map
	notifyEnabled bool
}

func NewMockAsyncer(notifyEnabled bool) *MockAsyncer {
	return &MockAsyncer{
		notifyEnabled: notifyEnabled,
	}
}

func (a *MockAsyncer) ContentType(key string) ContentType {
	if strings.HasSuffix(key, ".yml") {
		return T_YAML
	}

	return T_JSON
}

func (a *MockAsyncer) Get(key string) []byte {
	v, ok := a.data.Load(key)
	if !ok {
		return nil
	}

	vv := v.([]byte)

	var m map[string]interface{}
	mar := typeMarshalers[a.ContentType(key)]
	err := mar.Unmarshal(vv, &m)

	if err != nil {
		logger.Errorf("unmarshal err:%v", err)
		return nil
	}

	atomic.AddInt32(&a.ct, 1)
	setMapValue(m, "key", key)
	setMapValue(m, "ct", a.ct)
	ret, _ := mar.Marshal(m)
	logger.Infof("get async config[%s]:%s", key, ret)
	return ret
}

func (a *MockAsyncer) Set(key string, value []byte) error {
	a.data.Store(key, value)
	logger.Infof("set async config[%s]:%s", key, value)
	a.notify(key)
	return nil
}

func (a *MockAsyncer) notify(key string) {
	if !a.notifyEnabled {
		return
	}
	if ch, ok := a.notifyChans.Load(key); ok {
		ch.(chan struct{}) <- struct{}{}
	}
}

func (a *MockAsyncer) Watch(key string) chan struct{} {
	if !a.notifyEnabled {
		return nil
	}

	if ch, ok := a.notifyChans.Load(key); ok {
		return ch.(chan struct{})
	}

	ch := make(chan struct{}, 1)
	a.notifyChans.Store(key, ch)

	return ch
}
