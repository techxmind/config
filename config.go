package config

import (
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

const (
	DefaultLayerName    = "default"
	RootKey             = ""
	RemoteConfSourceKey = "remote_conf_source"
)

var (
	errEmptyKeyPath         = errors.New("empty key path")
	errKeyPathPrefixInvalid = errors.New("invalid key path prefix")
	errKeyPathIncorrect     = errors.New("incorrect  key path")
)

func init() {
}

type Config struct {
	layers            sync.Map //[string]Configer layerName => Configer
	proxyPool         sync.Pool
	defaultLayerNames atomic.Value //[]string
	ConfigHelper
}

type defaultConfig struct {
	cfg *Config
}

func (c *defaultConfig) Get(keyPath string) interface{} {
	return c.cfg.Get(keyPath)
}

func (c *defaultConfig) Set(keyPath string, value interface{}) error {
	return c.cfg.Set(keyPath, value)
}

func (c *defaultConfig) Watch(notifier chan bool) {
	c.cfg.Watch(notifier)
}

func newConfig() *Config {
	c := &Config{}
	c.ConfigHelper = ConfigHelper{
		Configer: &defaultConfig{
			cfg: c,
		},
	}
	c.proxyPool = sync.Pool{
		New: func() interface{} {
			return NewLayerConfigProxy(c)
		},
	}
	c.AddDefaultLayerName(DefaultLayerName)
	return c
}

// AddDefaultLayerName add new layer to default layer.
// The last one is the first searching path.
func (cfg *Config) AddDefaultLayerName(layerName string) {
	origins, _ := cfg.defaultLayerNames.Load().([]string)
	m := make(map[string]bool)
	s := make([]string, 0, len(origins)+1)
	s = append(s, layerName)
	m[layerName] = true
	for _, item := range origins {
		if _, ok := m[item]; !ok {
			m[item] = true
			s = append(s, item)
		}
	}
	cfg.defaultLayerNames.Store(s)
}

func AddDefaultLayerName(layerName string) {
	_cfg.AddDefaultLayerName(layerName)
}

func (cfg *Config) RemoveDefaultLayerName(layerName string) {
	origins, _ := cfg.defaultLayerNames.Load().([]string)
	news := make([]string, 0, len(origins))
	for _, item := range origins {
		if item != layerName {
			news = append(news, item)
		}
	}
	cfg.defaultLayerNames.Store(news)
}

func RemoveDefaultLayerName(layerName string) {
	_cfg.RemoveDefaultLayerName(layerName)
}

func (cfg *Config) AddLayer(layerName string, layer Configer) {
	cfg.layers.Store(layerName, layer)
}

func AddLayer(layerName string, layer Configer) {
	_cfg.AddLayer(layerName, layer)
}

func (cfg *Config) RemoteLayer(path string, remoteSources ...string) (*LayerConfigProxy, error) {
	if _, ok := cfg.layers.Load(path); !ok {
		remoteSource := cfg.StringDefault(RemoteConfSourceKey, "redis")
		if len(remoteSources) > 0 {
			remoteSource = remoteSources[0]
		}
		args := GetAsyncer(remoteSource)
		if args != nil {
			layer := NewAsyncConfig(args.Ins, path, args.CacheTime, args.RefreshAsync)
			cfg.AddLayer(path, layer)
		} else {
			return nil, errors.Errorf("unsupport remote source[%s], maybe config not inited?", remoteSource)
		}
	}

	return cfg.Layer(path), nil
}

func RemoteLayer(path string, remoteSources ...string) (*LayerConfigProxy, error) {
	return _cfg.RemoteLayer(path, remoteSources...)
}

func (cfg *Config) RemoveLayer(layerName string) {
	cfg.layers.Delete(layerName)
}

func RemoveLayer(layerName string) {
	_cfg.RemoveLayer(layerName)
}

// 获取Layer访问的代理对象
//
//  layer := cfg.Layer("layer1", "layer2")
//  layer.String("config_keyPath_from_layer1_or_layer2")
//  //等价于上面的调用 cfg.String("config_keyPath_from_layer1_or_layer2", "layer1", "layer2")
//
func (cfg *Config) Layer(layerNames ...string) *LayerConfigProxy {
	proxy := cfg.proxyPool.Get().(*LayerConfigProxy)
	proxy.SetLayerNames(layerNames...)
	return proxy
}

func Layer(layerNames ...string) *LayerConfigProxy {
	return _cfg.Layer(layerNames...)
}

// 归还proxy对象，方便后续复用
func (cfg *Config) PutLayer(p *LayerConfigProxy) {
	if p != nil {
		cfg.proxyPool.Put(p)
	}
}

func PutLayer(p *LayerConfigProxy) {
	_cfg.PutLayer(p)
}

// 从指定的Layer中获取配置值，未指定LayerNames，默认为DefaultLayerNames
//
//  cfg.Get("service_url") // Same of cfg.Get("service_url", DefaultLayerName)
//  cfg.Get("service_url", DefaultLayerName, "billing") // 尝试依次从默认配置，"billing"配置中查询service_url的配置
//
func (cfg *Config) Get(keyPath string, layerNames ...string) (val interface{}) {
	if len(layerNames) == 0 {
		layerNames = cfg.defaultLayerNames.Load().([]string)
	}

	for _, layerName := range layerNames {
		if layer, ok := cfg.layers.Load(layerName); ok {
			val = layer.(Configer).Get(keyPath)
			if val != nil {
				break
			}
		}
	}

	return
}

func Get(keyPath string, layerNames ...string) (val interface{}) {
	return _cfg.Get(keyPath, layerNames...)
}

func (cfg *Config) Watch(notifier chan bool, layerNames ...string) {
	if len(layerNames) == 0 {
		layerNames = cfg.defaultLayerNames.Load().([]string)
	}

	for _, layerName := range layerNames {
		if layer, ok := cfg.layers.Load(layerName); ok {
			layer.(Configer).Watch(notifier)
		}
	}
}

func Watch(notifier chan bool, layerNames ...string) {
	_cfg.Watch(notifier, layerNames...)
}

// 设置指定Layer的配置，LayerNames不传默认为DefaultLayerName
// 性能较低(359913 ns/op)：每次调会clone一个新的副本，并在副本上更新，替换原配置map
//
//  cfg.Set("db.host", "example.com")
//  cfg.Set(RootKey, map[string]interface{}{
//    "db" : map[string]interface{}{
//      "host" : "example.com",
//     },
//  })
//
func (cfg *Config) Set(keyPath string, value interface{}, layerNames ...string) error {

	if value == nil {
		return nil
	}

	if len(layerNames) == 0 {
		layerNames = []string{DefaultLayerName}
	}

	layerName := layerNames[0]

	if layer, ok := cfg.layers.Load(layerName); ok {
		return layer.(Configer).Set(keyPath, value)
	}

	return errors.Errorf("set config error, layer[%s] not exist", layerName)
}

func Set(keyPath string, value interface{}, layerNames ...string) error {
	return _cfg.Set(keyPath, value, layerNames...)
}

func Merge(value interface{}, layerNames ...string) error {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.Merge(value)
}

func Map(keyPath string, layerNames ...string) *MapConfig {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.Map(keyPath)
}

func JSON(keyPath string, layerNames ...string) ([]byte, error) {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.JSON(keyPath)
}

func Remarshal(keyPath string, v interface{}, layerNames ...string) error {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.Remarshal(keyPath, v)
}

func String(keyPath string, layerNames ...string) string {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.String(keyPath)
}

func StringDefault(keyPath string, dft string, layerNames ...string) (value string) {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.StringDefault(keyPath, dft)
}

func Bytes(keyPath string, layerNames ...string) []byte {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.Bytes(keyPath)
}

func BytesDefault(keyPath string, dft []byte, layerNames ...string) (value []byte) {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.BytesDefault(keyPath, dft)
}

func Float(keyPath string, layerNames ...string) float64 {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.Float(keyPath)
}

func FloatDefault(keyPath string, dft float64, layerNames ...string) float64 {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.FloatDefault(keyPath, dft)
}

func Int(keyPath string, layerNames ...string) int64 {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.Int(keyPath)
}

func IntDefault(keyPath string, dft int64, layerNames ...string) int64 {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.IntDefault(keyPath, dft)
}

func Uint(keyPath string, layerNames ...string) uint64 {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.Uint(keyPath)
}

func UintDefault(keyPath string, dft uint64, layerNames ...string) uint64 {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.UintDefault(keyPath, dft)
}

func Bool(keyPath string, layerNames ...string) bool {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.Bool(keyPath)
}

func BoolDefault(keyPath string, dft bool, layerNames ...string) bool {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.BoolDefault(keyPath, dft)
}
