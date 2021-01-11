package config

import (
	"encoding/json"
	"sync"
	"sync/atomic"
)

type MockAsyncer struct {
	ct   int32
	data sync.Map
}

func NewMockAsyncer() *MockAsyncer {
	return &MockAsyncer{}
}

func (a *MockAsyncer) Get(keyPath string) []byte {
	v, ok := a.data.Load(keyPath)
	if !ok {
		return nil
	}

	vv := v.([]byte)

	var m map[string]interface{}
	err := json.Unmarshal(vv, &m)

	if err != nil {
		logger.Errorf("err:%v", err)
		return nil
	}

	atomic.AddInt32(&a.ct, 1)
	setMapValue(m, "key", keyPath)
	setMapValue(m, "ct", a.ct)
	ret, _ := json.Marshal(m)
	logger.Infof("get async config[%s]:%s", keyPath, ret)
	return ret
}

func (a *MockAsyncer) Set(keyPath string, value []byte) error {
	a.data.Store(keyPath, value)
	logger.Infof("set async config[%s]:%s", keyPath, value)
	return nil
}

func (a *MockAsyncer) Watch(keyPath string) chan bool {
	return nil
}
