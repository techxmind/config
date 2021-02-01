package config

type Configer interface {
	Get(keyPath string) interface{}
	Set(keyPath string, value interface{}) error
	Watch(notifier chan struct{})
}

type Config interface {
	Configer
	JSON(keyPath string) ([]byte, error)
	Remarshal(keyPath string, v interface{}) error
	Dump(keyPath string)
	Map(keyPath string) *MapConfig
	Merge(value interface{}) error
	String(keyPath string) string
	StringDefault(keyPath string, dft string) (value string)
	Bytes(keyPath string) []byte
	BytesDefault(keyPath string, dft []byte) (value []byte)
	Float(keyPath string) float64
	FloatDefault(keyPath string, dft float64) float64
	Int(keyPath string) int64
	IntDefault(keyPath string, dft int64) int64
	Uint(keyPath string) uint64
	UintDefault(keyPath string, dft uint64) uint64
	Bool(keyPath string) bool
	BoolDefault(keyPath string, dft bool) bool
}
