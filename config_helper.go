package config

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/techxmind/go-utils/itype"
)

type ConfigHelper struct {
	Configer
}

// JSON 返回指定节点的JSON数据
//
func (h *ConfigHelper) JSON(keyPath string) ([]byte, error) {
	val := h.Get(keyPath)

	if val == nil {
		return nil, errors.Errorf("path[%s] is nil", keyPath)
	}

	return json.Marshal(val)
}

// Remarshal 指定配置重新unmarshal为v
//
func (h *ConfigHelper) Remarshal(keyPath string, v interface{}) error {
	bs, err := h.JSON(keyPath)

	if err != nil {
		return err
	}

	return json.Unmarshal(bs, v)
}

// Dump 打印指定节点的配置JSON
//
func (h *ConfigHelper) Dump(keyPath string) {
	PrintJSON(h.Get(keyPath))
}

// Map 返回子配置
//
//  m := cfg.Map("key")
//  val1 := m.String("val1") // same as cfg.String("key.val1")
//  val2 := m.String("val2.val22") // same as cfg.String("key.val2.val22")
func (h *ConfigHelper) Map(keyPath string) *MapConfig {
	val := h.Get(keyPath)

	if val == nil {
		return nil
	}

	if m, ok := val.(map[string]interface{}); ok {
		return NewMapConfig(m)
	}

	return nil
}

// Merge 合并配置
//
// alias to Set(RootKey ...)
func (h *ConfigHelper) Merge(value interface{}) error {
	return h.Set(RootKey, value)
}

// String 返回指定节点string类型的配置值
//
func (h *ConfigHelper) String(keyPath string) string {

	return itype.String(h.Get(keyPath))
}

// StringDefault 返回指定节点string类型的配置值，不存在则返回默认值
//
func (h *ConfigHelper) StringDefault(keyPath string, dft string) (value string) {
	value = h.String(keyPath)

	if value == "" {
		value = dft
	}

	return
}

// Bytes 返回指定节点[]byte类型的配置值
//
func (h *ConfigHelper) Bytes(keyPath string) []byte {
	return itype.Bytes(h.Get(keyPath))
}

// BytesDefault 返回指定节点[]byte类型的配置值，不存在则返回默认值
//
func (h *ConfigHelper) BytesDefault(keyPath string, dft []byte) (value []byte) {
	if value = h.Bytes(keyPath); len(value) == 0 {
		value = dft
	}

	return
}

// Float 返回指定节点float64类型的配置值
//
func (h *ConfigHelper) Float(keyPath string) float64 {

	return itype.Float(h.Get(keyPath))
}

// FloatDefault 返回指定节点float64类型的配置值，不存在则返回默认值
//
func (h *ConfigHelper) FloatDefault(keyPath string, dft float64) float64 {
	ivalue := h.Get(keyPath)
	if ivalue == nil {
		return dft
	}

	return itype.Float(ivalue)
}

// Int 返回指定节点int64类型的配置值
//
func (h *ConfigHelper) Int(keyPath string) int64 {

	return itype.Int(h.Get(keyPath))
}

// IntDefault 返回指定节点int64类型的配置值，不存在则返回默认值
//
func (h *ConfigHelper) IntDefault(keyPath string, dft int64) int64 {
	ivalue := h.Get(keyPath)
	if ivalue == nil {
		return dft
	}

	return itype.Int(ivalue)
}

// Uint 返回指定节点uint64类型的配置值
//
func (h *ConfigHelper) Uint(keyPath string) uint64 {

	return itype.Uint(h.Get(keyPath))
}

// UintDefault 返回指定节点uint64类型的配置值，不存在则返回默认值
//
func (h *ConfigHelper) UintDefault(keyPath string, dft uint64) uint64 {
	ivalue := h.Get(keyPath)
	if ivalue == nil {
		return dft
	}

	return itype.Uint(ivalue)
}

// Bool 返回指定节点bool类型的配置值
//
func (h *ConfigHelper) Bool(keyPath string) bool {
	str := itype.String(h.Get(keyPath))

	return !isFalseStr(str)
}

// BoolDefault 返回指定节点bool类型的配置值，不存在则返回默认值
//
func (h *ConfigHelper) BoolDefault(keyPath string, dft bool) bool {
	ivalue := h.Get(keyPath)
	if ivalue == nil {
		return dft
	}

	return !isFalseStr(itype.String(ivalue))
}
