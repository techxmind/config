package config

// config method proxy for auto pass layerNames parameter
//
//  p := NewLayerConfigProxy(cfg, "layer1", "layer2")
//  p.String("key") // alias to : cfg.String("key", "layer1", "layer2")
//
type LayerConfigProxy struct {
	ConfigHelper
}

func NewLayerConfigProxy(cfg *defaultConfig, layerNames ...string) *LayerConfigProxy {
	p := &layerConfigProxy{
		layerNames: layerNames,
		cfg:        cfg,
	}

	return &LayerConfigProxy{
		ConfigHelper: ConfigHelper{
			Configer: p,
		},
	}
}

// not thread-safe
func (p *LayerConfigProxy) SetLayerNames(layerNames ...string) {
	p.ConfigHelper.Configer.(*layerConfigProxy).layerNames = layerNames
}

type layerConfigProxy struct {
	layerNames []string
	cfg        *defaultConfig
}

func (p *layerConfigProxy) Get(keyPath string) (val interface{}) {
	return p.cfg.Get2(keyPath, p.layerNames...)
}

func (p *layerConfigProxy) Set(keyPath string, value interface{}) error {
	return p.cfg.Set2(keyPath, value, p.layerNames...)
}

func (p *layerConfigProxy) Watch(notifier chan struct{}) {
	p.cfg.Watch2(notifier, p.layerNames...)
}
