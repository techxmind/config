package config

import (
	"flag"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/techxmind/go-utils/fileutil"
)

var (
	_cfg *defaultConfig

	_opts = struct {
		// Value cache time(seconds) when asyncer do not support value changed notify.
		cacheTime int

		// When value is expired (check triggered by request), refresh the value asynchronously.
		// It will cause the request that triggered the refresh get the old value.
		refreshAsync bool

		// file config
		file      string
		fileAlive bool

		// redis config
		redisAddr       string
		redisDb         int
		redisPassword   string
		redisDefaultKey string
		redisSubChannel string
	}{
		cacheTime: 3,
	}

	DefaultRedisAsyncer *RedisAsyncer
)

func init() {
	opts := []struct {
		v            interface{}
		vType        string
		name         string
		defaultValue interface{}
		desc         string
	}{
		{&_opts.file, "string", "conf.file", "", "Config file path"},
		{&_opts.fileAlive, "bool", "conf.file.alive", false, "Reload config file when content changed"},
		{&_opts.redisAddr, "string", "conf.redis.addr", "", "Redis address that connected to get config"},
		{&_opts.redisPassword, "string", "conf.redis.password", "", "Redis password"},
		{&_opts.redisDb, "int", "conf.redis.db", 0, "Redis db number"},
		{&_opts.redisDefaultKey, "string", "conf.redis.default", "", "Redis default key that contains default config"},
		{&_opts.redisSubChannel, "string", "conf.redis.channel", "", "Redis channel to subscribe value changed event"},
		{&_opts.cacheTime, "int", "conf.cache_time", 3, "Value cache time(seconds) when asyncer do not support value changed notify"},
		{&_opts.refreshAsync, "bool", "conf.refresh_async", false, "Refresh value asynchronously or not"},
	}

	for _, opt := range opts {
		if opt.vType == "string" {
			flag.StringVar(opt.v.(*string), opt.name, opt.defaultValue.(string), opt.desc)
		} else if opt.vType == "int" {
			flag.IntVar(opt.v.(*int), opt.name, opt.defaultValue.(int), opt.desc)
		} else if opt.vType == "bool" {
			flag.BoolVar(opt.v.(*bool), opt.name, opt.defaultValue.(bool), opt.desc)
		}
	}

	_cfg = newConfig()
	_cfg.AddLayer(DefaultLayerName, NewMapConfig(make(map[string]interface{})))

	// use local FlagSet to call parse method immediately
	logArgs := regexp.MustCompile(`-{1,2}(?:conf(?:\.[\w.]+)?)(?:\s+|\s*=\s*)(?:\S+)`).
		FindAllString(strings.Join(os.Args[1:], " "), -1)
	if len(logArgs) > 0 {
		flagSet := flag.NewFlagSet("config", flag.ContinueOnError)
		for _, opt := range opts {
			if opt.vType == "string" {
				flagSet.StringVar(opt.v.(*string), opt.name, opt.defaultValue.(string), opt.desc)
			} else if opt.vType == "int" {
				flagSet.IntVar(opt.v.(*int), opt.name, opt.defaultValue.(int), opt.desc)
			} else if opt.vType == "bool" {
				flagSet.BoolVar(opt.v.(*bool), opt.name, opt.defaultValue.(bool), opt.desc)
			}
		}
		r := regexp.MustCompile(`\s+`)
		args := make([]string, 0, len(logArgs))
		for _, arg := range logArgs {
			pairs := r.Split(arg, 2)
			args = append(args, pairs...)
		}
		flagSet.Parse(args)
	}

	cacheTime := time.Duration(_opts.cacheTime) * time.Second

	if _opts.file != "" {
		initWithFile(
			_opts.file,
			_opts.fileAlive,
			cacheTime,
			_opts.refreshAsync,
		)
	}

	if _opts.redisAddr != "" {
		initWithRedis(
			&redis.Options{
				Addr:     _opts.redisAddr,
				Password: _opts.redisPassword,
				DB:       _opts.redisDb,
			},
			_opts.redisSubChannel,
			_opts.redisDefaultKey,
			cacheTime,
			_opts.refreshAsync,
		)
	}
}

// initWithRedis load config from redis and set it to default layer
//
func initWithRedis(redisOpts *redis.Options, channel string, defaultKey string, cacheTime time.Duration, refreshAsync bool) {
	DefaultRedisAsyncer = NewRedisAsyncer(redisOpts, channel)

	RegisterAsyner("redis", &AsyncerArgs{
		Ins:          DefaultRedisAsyncer,
		CacheTime:    cacheTime,
		RefreshAsync: refreshAsync,
	})

	if defaultKey == "" {
		return
	}

	redisCfg := NewAsyncConfig(
		DefaultRedisAsyncer,
		defaultKey,
		cacheTime,
		refreshAsync,
	)

	layerName := "default-conf-redis"
	_cfg.AddLayer(layerName, redisCfg)
	AddDefaultLayerName(layerName)
}

// initConfFromFile load config from file and set it to default layer
//
func initWithFile(file string, alive bool, cacheTime time.Duration, refreshAsync bool) {
	if file == "" {
		logger.Errorf("conf file unspecified")
		return
	}

	if !fileutil.Exist(file) {
		logger.Fatalf("conf file[%s] not found", file)
	}

	fileCfg := NewAsyncConfig(NewFileAsyncer(), file, cacheTime, refreshAsync)

	if !alive {
		// 静态配置文件，直接合并至默认层，提高配置查询的性能
		_cfg.Merge(fileCfg.Get(RootKey))
	} else {
		layerName := "default-conf-file"
		_cfg.AddLayer(layerName, fileCfg)
		AddDefaultLayerName(layerName)
	}
}
