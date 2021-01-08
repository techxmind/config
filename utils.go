package config

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/techxmind/go-utils/object"
)

func JSONToMap(rawMessage json.RawMessage) (map[string]interface{}, error) {
	var data interface{}

	json.Unmarshal(rawMessage, &data)

	if m, ok := data.(map[string]interface{}); ok {
		return m, nil
	}

	return nil, fmt.Errorf("json data is not map[string]interface{} struct")
}

func PrintJSON(v interface{}) {
	json, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(json))
}

func isFalseStr(str string) bool {
	str = strings.ToLower(str)

	return str == "" || str == "0" || str == "false" || str == "f" || str == "off"
}

func setMapValue(m map[string]interface{}, keyPath string, value interface{}) error {

	keys := strings.Split(keyPath, ".")
	lastKey := keys[len(keys)-1]

	var obj interface{}
	if len(keys) > 1 {
		obj, _ = object.GetObject(m, strings.Join(keys[:len(keys)-1], "."), true)
	} else {
		obj = m
	}

	if obj == nil {
		return fmt.Errorf("config[%s] type error", keyPath)
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		// 设置某key为nil，则删除该key
		if value == nil {
			delete(v, lastKey)
		} else {
			v[lastKey] = value
		}
	case []interface{}:
		if index, err := strconv.ParseInt(lastKey, 10, 32); err == nil {
			if int(index) >= len(v) {
				return fmt.Errorf("config[%s] array out of range[%d]", keyPath, index)
			}
			v[int(index)] = value
		} else {
			return err
		}
	default:
		return fmt.Errorf("config[%s] type error", keyPath)
	}

	return nil
}

// 合并两个Map, 当子节点都为Map时，会深度合并, 否则新值会覆盖旧值
func mergeMap(originMap map[string]interface{}, extraMap map[string]interface{}) {
	for k, v := range extraMap {
		originV, originExist := originMap[k]
		if !originExist {
			originMap[k] = v
			continue
		}
		originSubMap, originSubMapOk := originV.(map[string]interface{})
		if !originSubMapOk {
			originMap[k] = v
			continue
		}
		if subMap, ok := v.(map[string]interface{}); ok {
			mergeMap(originSubMap, subMap)
		} else {
			originMap[k] = v
			continue
		}
	}
}

func dump(vals ...interface{}) {
	fmt.Println(vals...)
}
