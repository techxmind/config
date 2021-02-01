package config

import (
	"crypto/md5"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"
	"golang.org/x/sync/singleflight"

	"github.com/techxmind/go-utils/object"
)

var (
	_asyncers sync.Map

	// for mock
	_now = time.Now
)

type AsyncerArgs struct {
	Ins          Asyncer
	CacheTime    time.Duration
	RefreshAsync bool
}

func RegisterAsyner(typeName string, args *AsyncerArgs) {
	_asyncers.Store(typeName, args)
}

func GetAsyncer(typeName string) *AsyncerArgs {
	args, ok := _asyncers.Load(typeName)
	if !ok {
		return nil
	}

	return args.(*AsyncerArgs)
}

type Asyncer interface {
	ContentType(key string) ContentType
	Get(key string) []byte
	Set(key string, value []byte) error
	Watch(key string) chan struct{} // 实时监控配置变化
}

// 远程配置 qconf/consul/database
type AsyncConfig struct {
	ConfigHelper
}

// NewAsyncConfig 异步配置（qconf/consul/rds...）
//
// asyncer: 实现异步获取及设置接口的对象
// asyncKey: 获取整个异步的Key（和Get方法的Key要区分）
// cacheTime: 配置缓存的时间，超过该缓存时间会触发重新获取异步数据. <= 0 数据不过期
// refreshAsync: 缓存过期时，刷新数据是同步还是异步（同步：有查询请求时，会等待数据刷新完成，异步则不会等待）
func NewAsyncConfig(asyncer Asyncer, asyncKey string, cacheTime time.Duration, refreshAsync bool) *AsyncConfig {
	contentType := asyncer.ContentType(asyncKey)

	cfg := &asyncConfig{
		asyncKey:     asyncKey,
		marshaler:    typeMarshalers[contentType],
		contentType:  contentType,
		asyncer:      asyncer,
		cacheTime:    cacheTime,
		refreshAsync: refreshAsync,
		quit:         make(chan struct{}),
	}

	cfg.refresh()

	if notify := asyncer.Watch(asyncKey); notify != nil {
		// 推送更新机制下可以不使用过期策略
		// 但为了防止更新消息丢失导致的旧值一直得不到更新
		// 设置一个兜底的过期时间
		cfg.cacheTime = 5 * time.Minute
		go cfg.watch(notify)
	}

	return &AsyncConfig{
		ConfigHelper: ConfigHelper{
			Configer: cfg,
		},
	}
}

type asyncConfig struct {
	sync.Mutex
	asyncKey      string
	marshaler     Marshaler
	contentType   ContentType
	value         atomic.Value
	rawMessageMd5 string

	sf singleflight.Group

	notifiers []chan struct{}

	asyncer      Asyncer
	refreshAsync bool
	refreshTime  int64
	cacheTime    time.Duration
	quit         chan struct{}
}

func (cfg *asyncConfig) watch(notify chan struct{}) {
	for {
		select {
		case <-notify:
			cfg.refresh()

		//TODO
		case <-cfg.quit:
			return
		}
	}
}

func (cfg *asyncConfig) Get(keyPath string) interface{} {
	now := _now().UnixNano()
	refreshTime := atomic.LoadInt64(&cfg.refreshTime)
	if cfg.cacheTime > 0 && time.Duration(now-refreshTime)*time.Nanosecond > cfg.cacheTime { // content expired
		if refreshTime > 0 && cfg.refreshAsync { // if the content initialized and refreshAsync setted
			logger.Debugf("asyncer[%s] refresh async", cfg.asyncKey)
			go cfg.refresh()
		} else { // 同步更新
			logger.Debugf("asyncer[%s] refresh sync, cacheTime=%d, refreshTime=%d", cfg.asyncKey, cfg.cacheTime, refreshTime)
			cfg.refresh()
		}
	}

	if keyPath == RootKey {
		return cfg.value.Load()
	}

	val, ok := object.GetValue(cfg.value.Load(), keyPath)
	if !ok {
		return nil
	}

	return val
}

func (cfg *asyncConfig) refresh() {
	cfg.sf.Do("", func() (_ interface{}, _ error) {
		atomic.StoreInt64(&cfg.refreshTime, _now().UnixNano())

		rawMessage := cfg.asyncer.Get(cfg.asyncKey)
		rawMessage = processRawMessage(rawMessage, cfg.contentType)

		if len(rawMessage) == 0 {
			logger.Warnf("asyncer[%s] get empty content", cfg.asyncKey)
			return
		}

		rawMessageMd5 := fmt.Sprintf("%x", md5.Sum(rawMessage))

		// no change
		if rawMessageMd5 == cfg.rawMessageMd5 {
			return
		}

		var val interface{}
		if err := cfg.marshaler.Unmarshal(rawMessage, &val); err != nil {
			logger.Errorf("unmarshal async config[%s] error:%v", cfg.asyncKey, err)
			return
		}
		cfg.rawMessageMd5 = rawMessageMd5
		cfg.value.Store(val)

		cfg.notify()

		return
	})
}

// Set 设置配置
//
// 注意：配置自动刷新会覆盖手动设置的同名配置值
func (cfg *asyncConfig) Set(keyPath string, value interface{}) error {
	cfg.Lock()
	defer cfg.Unlock()

	var iorigin interface{}

	if keyPath == RootKey {
		cfg.value.Store(value)
	} else {
		iorigin = cfg.value.Load()
		if iorigin == nil {
			iorigin = make(map[string]interface{})
		}
		origin, ok := iorigin.(map[string]interface{})
		if !ok {
			return errors.Errorf("Set config[%s] %s=%v error", cfg.asyncKey, keyPath, value)
		}
		newValue := deepcopy.Copy(origin).(map[string]interface{})
		if err := setMapValue(newValue, keyPath, value); err != nil {
			return err
		}
		cfg.value.Store(newValue)
	}

	data, err := cfg.marshaler.Marshal(cfg.value.Load())
	if err != nil {
		return err
	}

	cfg.notify()

	return cfg.asyncer.Set(cfg.asyncKey, data)
}

func (cfg *asyncConfig) notify() {
	for _, notifier := range cfg.notifiers {
		select {
		case notifier <- struct{}{}:
		default:
		}
	}
}

func (cfg *asyncConfig) Watch(notifier chan struct{}) {
	cfg.Lock()
	defer cfg.Unlock()
	cfg.notifiers = append(cfg.notifiers, notifier)
}
