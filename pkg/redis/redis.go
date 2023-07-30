package redis

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	_defaultAddress      = "localhost:6379"
	_defaultPassword     = ""
	_defaultReadTimeout  = 3 * time.Second
	_defaultWriteTimeout = 3 * time.Second
	_defaultDB           = 0
	_defaultMinIdleConn  = 0
	_defaultMaxIdleConn  = 5
	_defaultMaxIdleTime  = 5 * time.Minute
)

var (
	once                      sync.Once
	redisClientSingleInstance *RedisClient
)

type RedisClient struct {
	host         string
	password     string
	db           int
	readTimeout  time.Duration
	writeTimeout time.Duration

	minIdleConn int
	maxIdleConn int
	maxIdleTime time.Duration

	Client *redis.Client
}

func NewRedisClient(opts ...Option) *RedisClient {
	if redisClientSingleInstance == nil {
		once.Do(func() {
			redisClientSingleInstance = &RedisClient{
				host:         _defaultAddress,
				password:     _defaultPassword,
				readTimeout:  _defaultReadTimeout,
				writeTimeout: _defaultWriteTimeout,
				minIdleConn:  _defaultMinIdleConn,
				maxIdleConn:  _defaultMaxIdleConn,
				maxIdleTime:  _defaultMaxIdleTime,
			}

			for _, opt := range opts {
				opt(redisClientSingleInstance)
			}

			instance := redis.NewClient(&redis.Options{
				Addr:                  redisClientSingleInstance.host,
				ClientName:            "SinarLog",
				Password:              redisClientSingleInstance.password,
				DB:                    redisClientSingleInstance.db,
				ReadTimeout:           redisClientSingleInstance.readTimeout,
				WriteTimeout:          redisClientSingleInstance.writeTimeout,
				ContextTimeoutEnabled: true,
				MinIdleConns:          redisClientSingleInstance.minIdleConn,
				MaxIdleConns:          redisClientSingleInstance.maxIdleConn,
				ConnMaxIdleTime:       redisClientSingleInstance.maxIdleTime,
			})

			if err := instance.Ping(context.Background()).Err(); err != nil {
				log.Fatalf("unable to connect to redis: %s", err.Error())
			}

			redisClientSingleInstance.Client = instance
		})
	}

	return redisClientSingleInstance
}
