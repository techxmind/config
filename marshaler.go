package config

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// config raw content type
type ContentType int

const (
	T_JSON ContentType = iota
	T_YAML
)

var (
	typeMarshalers = map[ContentType]Marshaler{
		T_JSON: JSONMarshaler{},
		T_YAML: YAMLMarshaler{},
	}
)

type Marshaler interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
}

// JSONMarshaler
type JSONMarshaler struct{}

func (m JSONMarshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (m JSONMarshaler) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// YAMLMarshaler
type YAMLMarshaler struct{}

func (m YAMLMarshaler) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func (m YAMLMarshaler) Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}
