package config

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
)

type RedisAsyncer struct {
	db            *redis.Client
	ctx           context.Context
	notifyEnabled bool
	notifyChans   sync.Map
}

// NewRedisAsyncer create new RedisAsyncer.
// subChannel is channel name for subscribing to notify value changed,
// the notify feature will be disabled when subChannel is not specified.
func NewRedisAsyncer(options *redis.Options, subChannel string) *RedisAsyncer {
	db := redis.NewClient(options)
	a := &RedisAsyncer{
		db:  db,
		ctx: context.Background(),
	}

	if subChannel != "" {
		a.subscribe(subChannel)
	}

	logger.Infof("NewRedisAsyncer:subChannel=%s", subChannel)

	return a
}

func (a *RedisAsyncer) ContentType(key string) ContentType {
	// TODO: support yaml

	return T_JSON
}

func (a *RedisAsyncer) subscribe(channel string) {
	sub := a.db.Subscribe(a.ctx, channel)
	_, err := sub.Receive(a.ctx)
	if err != nil {
		logger.Errorf("redis subscribe channel=%s err=%v", channel, err)
		return
	}

	a.notifyEnabled = true
	go func() {
		for msg := range sub.Channel() {
			updatedKey := msg.Payload
			logger.Debugf("redis key updated:%s", updatedKey)
			a.notify(updatedKey)
		}
	}()
}

func (a *RedisAsyncer) Get(key string) []byte {
	val, err := a.db.Get(a.ctx, key).Result()

	if err == redis.Nil {
		return nil
	} else if err != nil {
		logger.Errorf("read conf[%s] from redis err:%v", key, err)
		return nil
	}

	if val == "" {
		return nil
	}

	bs := []byte(val)

	return bs
}

func (a *RedisAsyncer) Set(key string, content []byte) error {
	err := a.db.Set(context.Background(), key, string(content), 0).Err()

	if err == nil {
		a.notify(key)
	}

	return err
}

func (a *RedisAsyncer) notify(key string) {
	if a.notifyEnabled {
		if ch, ok := a.notifyChans.Load(key); ok {
			logger.Debugf("%s changed notify", key)
			ch.(chan struct{}) <- struct{}{}
		}
	}
}

func (a *RedisAsyncer) Watch(key string) chan struct{} {
	if !a.notifyEnabled {
		return nil
	}

	if ch, ok := a.notifyChans.Load(key); ok {
		return ch.(chan struct{})
	}

	ch := make(chan struct{}, 1)
	a.notifyChans.Store(key, ch)

	return ch
}
