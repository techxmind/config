package config

type Configer interface {
	Get(keyPath string) interface{}
	Set(keyPath string, value interface{}) error
	Watch(notifier chan bool)
}
