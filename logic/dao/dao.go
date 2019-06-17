package dao

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"github.com/swanky2009/goim/logic/g/conf"
	kafka "gopkg.in/Shopify/sarama.v1"
)

// Dao dao.
type Dao struct {
	c        *conf.Config
	kafkaPub kafka.SyncProducer
	redis    *redis.ClusterClient
	// redis       *redis.Client
	redisExpire time.Duration
}

// New new a dao and return.
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:           c,
		kafkaPub:    newKafkaPub(c.Kafka),
		redis:       newRedis(c.Redis),
		redisExpire: time.Duration(c.Redis.Expire),
	}
	return d
}

func newKafkaPub(c *conf.Kafka) kafka.SyncProducer {
	var err error
	kc := kafka.NewConfig()
	kc.Producer.RequiredAcks = kafka.WaitForAll // Wait for all in-sync replicas to ack the message
	kc.Producer.Retry.Max = 10                  // Retry up to 10 times to produce the message
	kc.Producer.Return.Successes = true
	pub, err := kafka.NewSyncProducer(c.Brokers, kc)
	if err != nil {
		panic(err)
	}
	return pub
}

func newRedis(c *conf.Redis) *redis.ClusterClient {
	redisdb := redis.NewClusterClient(
		// redisdb := redis.NewClient(
		&redis.ClusterOptions{
			// &redis.Options{
			Addrs:        c.Addrs,
			DialTimeout:  time.Duration(c.DialTimeout),
			ReadTimeout:  time.Duration(c.ReadTimeout),
			WriteTimeout: time.Duration(c.WriteTimeout),
			PoolSize:     c.PoolSize,
			PoolTimeout:  time.Duration(c.PoolTimeout),
			IdleTimeout:  time.Duration(c.IdleTimeout),
		})
	_, err := redisdb.Ping().Result()
	if err != nil {
		panic(err)
	}
	return redisdb
}

// Close close the resource.
func (d *Dao) Close() {
	d.redis.Close()
}

// Ping dao ping.
func (d *Dao) Ping(c context.Context) error {
	return d.pingRedis(c)
}
