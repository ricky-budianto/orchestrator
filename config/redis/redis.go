package redis

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	jsonExt "github.com/json-iterator/go"
	"github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/soluixdeveloper/ces-orchestratorService/config"
	"github.com/soluixdeveloper/ces-orchestratorService/config/utils"
)

var ctx = context.Background()

type cacheWriteOption func(*cacheWriteConfig) error

type cacheWriteConfig struct {
	// timeToLive berapa lama value akan disimpan di Redis
	timeToLive time.Duration
}

var (
	cachingRedis *redis.Client
)

func InitRedisClient() {
	if config.AppConfig.RedisUse {
		if config.RedisConfig.RedisWithoutPassword {
			cachingRedis = redis.NewClient(&redis.Options{
				Addr: config.RedisConfig.RedisURL + ":" + config.RedisConfig.RedisPort,
				DB:   0,
				TLSConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolSize:     10, // Adjust based on expected workload
				MinIdleConns: 3,  // Minimum idle connections
			})
		} else {
			cachingRedis = redis.NewClient(&redis.Options{
				Addr:     config.RedisConfig.RedisURL + ":" + config.RedisConfig.RedisPort,
				Password: config.RedisConfig.RedisPassword,
				DB:       0,
				TLSConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
				DialTimeout:  5 * time.Second,
				ReadTimeout:  3 * time.Second,
				WriteTimeout: 3 * time.Second,
				PoolSize:     10, // Adjust based on expected workload
				MinIdleConns: 3,  // Minimum idle connections
			})
		}
	}
}

func WriteCache(id, module string, data interface{}) error {
	return WriteCacheWith(id, module, data,
		WithTimeToLive(time.Duration(config.RedisConfig.CacheDuration)*time.Second),
	)
}

func WriteCacheWith(id, module string, data interface{}, options ...cacheWriteOption) error {
	logging := ceslogger.Logger{}

	if !config.AppConfig.RedisUse {
		logging.LogInfo("redis not used", utils.MessageNoCacheData)
		return nil
	}

	// Inisiasi config berdasar opsi yang diberikan
	cfg := cacheWriteConfig{}
	for _, opt := range options {
		err := opt(&cfg)
		if err != nil {
			return fmt.Errorf("failed to configure cache writing: %w", err)
		}
	}

	redisKey := config.AppConfig.ProjectName + "_" + module + "/" + id

	dataByte, _ := jsonExt.Marshal(data)

	err := cachingRedis.Set(ctx, redisKey, string(dataByte), cfg.timeToLive).Err()
	if err != nil {
		logging.LogError("cachingRedis.Set", err)
		return err
	}

	return err
}

func ReadCache(id, module string) ([]byte, error) {
	logging := ceslogger.Logger{}

	if !config.AppConfig.RedisUse {
		logging.LogInfo("redis not used", utils.MessageNoCacheData)
		return nil, nil
	}
	redisKey := config.AppConfig.ProjectName + "_" + module + "/" + id

	val, err := cachingRedis.Get(ctx, redisKey).Bytes()
	if err != nil {
		logging.LogError("cachingRedis.Get", err)
		return nil, err
	}
	return val, err
}

func SetWithoutExpiry(id, module string, data interface{}, exp time.Duration) error {
	logging := ceslogger.NewLogger("")

	if !config.AppConfig.RedisUse {
		logging.LogInfo("redis not used", utils.MessageNoCacheData)
		return nil
	}

	redisKey := config.AppConfig.ProjectName + "_" + module + "/" + id

	dataByte, _ := jsonExt.Marshal(data)

	err := cachingRedis.Set(ctx, redisKey, string(dataByte), exp*time.Second).Err()
	if err != nil {
		logging.LogError("cachingRedis.Set", err)
		return err
	}

	return err
}

// ~ write options ~
func WithTimeToLive(ttl time.Duration) cacheWriteOption {
	return func(cfg *cacheWriteConfig) error {
		cfg.timeToLive = ttl
		return nil
	}
}

func DeleteCache(id, module string) error {
	logging := ceslogger.NewLogger("")

	if !config.AppConfig.RedisUse {
		logging.LogInfo("redis not used", utils.MessageNoCacheData)
		return nil
	}
	redisKey := config.AppConfig.ProjectName + "_" + module + "/" + id

	err := cachingRedis.Del(ctx, redisKey).Err()
	if err != nil {
		logging.LogError("RedisData.Del", err)
		return err
	}
	return nil
}

func IsInitialized() bool {
	if !config.AppConfig.RedisUse {
		return true
	}
	return cachingRedis != nil
}

func Ping() error {
	if !config.AppConfig.RedisUse {
		return nil
	}
	_, err := cachingRedis.Ping(ctx).Result()
	return err
}

func WriteCacheNotExists(ctx context.Context, id, module string, value interface{}, timeToLive time.Duration) (bool, error) {
	logging := ceslogger.NewLogger(ctx.Value(utils.CorrelationIDKey).(string))

	if !config.AppConfig.RedisUse {
		logging.LogInfo("redis not used", utils.MessageNoCacheData)
		return false, errors.New("unavailable redis server")
	}
	redisKey := config.AppConfig.ProjectName + "_" + module + "/" + id

	result, err := cachingRedis.SetNX(ctx, redisKey, value, timeToLive).Result()
	if err != nil || !result {
		logging.LogError("cachingRedis set nx", err)
		return false, err
	}
	return result, err
}
