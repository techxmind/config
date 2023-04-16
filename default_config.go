package config

import (
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

const (
	DefaultLayerName     = "default"
	RootKey              = ""
	DefaultConfSourceKey = "default_conf_source"
)

var (
	errEmptyKeyPath         = errors.New("empty key path")
	errKeyPathPrefixInvalid = errors.New("invalid key path prefix")
	errKeyPathIncorrect     = errors.New("incorrect  key path")
)

func init() {
}

type defaultConfig struct {
	layers            sync.Map //[string]Configer layerName => Configer
	proxyPool         sync.Pool
	defaultLayerNames atomic.Value //[]string
	ConfigHelper
}

type defaultConfiger struct {
	cfg *defaultConfig
}

func (c *defaultConfiger) Get(keyPath string) interface{} {
	return c.cfg.Get2(keyPath)
}

func (c *defaultConfiger) Set(keyPath string, value interface{}) error {
	return c.cfg.Set2(keyPath, value)
}

func (c *defaultConfiger) Watch(notifier chan struct{}) {
	c.cfg.Watch2(notifier)
}

func newConfig() *defaultConfig {
	c := &defaultConfig{}
	c.ConfigHelper = ConfigHelper{
		Configer: &defaultConfiger{
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
func (cfg *defaultConfig) AddDefaultLayerName(layerName string) {
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

func (cfg *defaultConfig) RemoveDefaultLayerName(layerName string) {
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

func (cfg *defaultConfig) AddLayer(layerName string, layer Configer) {
	cfg.layers.Store(layerName, layer)
}

func AddLayer(layerName string, layer Configer) {
	_cfg.AddLayer(layerName, layer)
}

func (cfg *defaultConfig) Load(path string, sources ...string) (Config, error) {
	if _, ok := cfg.layers.Load(path); !ok {
		source := cfg.StringDefault(DefaultConfSourceKey, "redis")
		if len(sources) > 0 {
			source = sources[0]
		}
		args := GetAsyncer(source)
		if args != nil {
			layer := NewAsyncConfig(args.Ins, path, args.CacheTime, args.RefreshAsync)
			cfg.AddLayer(path, layer)
		} else {
			return nil, errors.Errorf("unsupport config source[%s], maybe config not inited?", source)
		}
	}

	return cfg.Layer(path), nil
}

func Load(path string, remoteSources ...string) (Config, error) {
	return _cfg.Load(path, remoteSources...)
}

func (cfg *defaultConfig) RemoveLayer(layerName string) {
	cfg.layers.Delete(layerName)
}

func RemoveLayer(layerName string) {
	_cfg.RemoveLayer(layerName)
}

// 获取Layer访问的代理对象
//
//	layer := cfg.Layer("layer1", "layer2")
//	layer.String("config_keyPath_from_layer1_or_layer2")
//	//等价于上面的调用 cfg.String("config_keyPath_from_layer1_or_layer2", "layer1", "layer2")
func (cfg *defaultConfig) Layer(layerNames ...string) *LayerConfigProxy {
	proxy := cfg.proxyPool.Get().(*LayerConfigProxy)
	proxy.SetLayerNames(layerNames...)
	return proxy
}

func Layer(layerNames ...string) *LayerConfigProxy {
	return _cfg.Layer(layerNames...)
}

// Default returns default layer config
func Default() *LayerConfigProxy {
	defaultNames, _ := _cfg.defaultLayerNames.Load().([]string)
	return Layer(defaultNames...)
}

// 归还proxy对象，方便后续复用
func (cfg *defaultConfig) PutLayer(p *LayerConfigProxy) {
	if p != nil {
		cfg.proxyPool.Put(p)
	}
}

func PutLayer(p *LayerConfigProxy) {
	_cfg.PutLayer(p)
}

// 从指定的Layer中获取配置值，未指定LayerNames，默认为DefaultLayerNames
//
//	cfg.Get2("service_url") // Same of cfg.Get("service_url", DefaultLayerName)
//	cfg.Get2("service_url", DefaultLayerName, "billing") // 尝试依次从默认配置，"billing"配置中查询service_url的配置
func (cfg *defaultConfig) Get2(keyPath string, layerNames ...string) (val interface{}) {
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
	return _cfg.Get2(keyPath, layerNames...)
}

func (cfg *defaultConfig) Watch2(notifier chan struct{}, layerNames ...string) {
	if len(layerNames) == 0 {
		layerNames = cfg.defaultLayerNames.Load().([]string)
	}

	for _, layerName := range layerNames {
		if layer, ok := cfg.layers.Load(layerName); ok {
			layer.(Configer).Watch(notifier)
		}
	}
}

func Watch(notifier chan struct{}, layerNames ...string) {
	_cfg.Watch2(notifier, layerNames...)
}

// 设置指定Layer的配置，LayerNames不传默认为DefaultLayerName
// 性能较低(359913 ns/op)：每次调会clone一个新的副本，并在副本上更新，替换原配置map
//
//	cfg.Set("db.host", "example.com")
//	cfg.Set(RootKey, map[string]interface{}{
//	  "db" : map[string]interface{}{
//	    "host" : "example.com",
//	   },
//	})
func (cfg *defaultConfig) Set2(keyPath string, value interface{}, layerNames ...string) error {

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
	return _cfg.Set2(keyPath, value, layerNames...)
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

func Exist(keyPath string, layerNames ...string) bool {
	p := _cfg.Layer(layerNames...)
	defer _cfg.PutLayer(p)
	return p.Exist(keyPath)
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
