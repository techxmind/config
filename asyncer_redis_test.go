package config

import (
	//"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	//"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type redisAsyncerTestSuite struct {
	suite.Suite
	rds           *miniredis.Miniredis
	notifyChannel string
	defaultKey    string
	defaultValue  string
}

func (s *redisAsyncerTestSuite) SetupSuite() {
	rds, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	s.rds = rds
}

func (s *redisAsyncerTestSuite) TearDownSuite() {
	s.rds.Close()
}

func (s *redisAsyncerTestSuite) SetupTest() {
	s.defaultKey = "default_config"
	s.notifyChannel = "value_changes"
	s.defaultValue = `
	{
		"foo" : {
			"bar" : 1, // comment
			"zap" : "zap_value"
		}
	}
	`
	s.rds.Set(s.defaultKey, s.defaultValue)
}

func (s *redisAsyncerTestSuite) TestBasic() {
	asyncer := NewRedisAsyncer(&redis.Options{
		Addr: s.rds.Addr(),
	}, "")

	s.EqualValues(s.defaultValue, asyncer.Get(s.defaultKey), "Asyncer.get")
	s.Nil(asyncer.Watch(s.defaultKey), "No watch channel")

	redisCfg := NewAsyncConfig(
		asyncer,
		s.defaultKey,
		5*time.Millisecond,
		false,
	)

	s.EqualValues(1, redisCfg.Int("foo.bar"), "get foo.bar")
	asyncer.Set(s.defaultKey, []byte(`
		{"foo" : { "bar" : 2 }}
	`))
	s.EqualValues(1, redisCfg.Int("foo.bar"), "get foo.bar")
	time.Sleep(6 * time.Millisecond)
	s.EqualValues(2, redisCfg.Int("foo.bar"), "get foo.bar")
}

func (s *redisAsyncerTestSuite) TestNotify() {
	asyncer := NewRedisAsyncer(&redis.Options{
		Addr: s.rds.Addr(),
	}, s.notifyChannel)

	s.NotNil(asyncer.Watch(s.defaultKey), "has watch channel")

	redisCfg := NewAsyncConfig(
		asyncer,
		s.defaultKey,
		5*time.Millisecond, // when notify enable, will ignore cache time
		false,
	)
	s.EqualValues(1, redisCfg.Int("foo.bar"), "get foo.bar")
	s.rds.Set(s.defaultKey, `
		{"foo" : { "bar" : 2 }}
	`)
	s.EqualValues(1, redisCfg.Int("foo.bar"), "get foo.bar")
	time.Sleep(6 * time.Millisecond)
	s.EqualValues(1, redisCfg.Int("foo.bar"), "get foo.bar")

	// trigger notify
	s.rds.Publish(s.notifyChannel, s.defaultKey)
	// wait update
	time.Sleep(1 * time.Millisecond)
	s.EqualValues(2, redisCfg.Int("foo.bar"), "get foo.bar")
}

func TestRedisAsyncerTestSuite(t *testing.T) {
	suite.Run(t, new(redisAsyncerTestSuite))
}
