package producer

import (
	"context"
	"mist/src/protos/v1/appuser"
	"mist/src/protos/v1/event"
	"time"

	"github.com/redis/go-redis/v9"
)

type MessageProducer interface {
	SendMessage(interface{}, event.ActionType, []*appuser.Appuser) error
}

type RedisInterface interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd
}

type MProducer struct {
	Redis RedisInterface
	Wp    *WorkerPool
}

type MProducerOptions struct {
	Workers     int
	ChannelSize int
}

func NewMProducer(redis RedisInterface) *MProducer {
	workers := 4
	queueSize := 100
	wp := NewWorkerPool(workers, queueSize)

	return &MProducer{Redis: redis, Wp: wp}
}

func NewMProducerOptions(redis RedisInterface, opts *MProducerOptions) *MProducer {
	wp := NewWorkerPool(opts.Workers, opts.ChannelSize)

	return &MProducer{Redis: redis, Wp: wp}
}

func (mp *MProducer) SendMessage(
	ctx context.Context, redisChannel string, data interface{}, action event.ActionType, appusers []*appuser.Appuser,
) {

	mp.Wp.AddJob(NewNotificationJob(ctx, redisChannel, data, action, appusers, mp.Redis))
}
