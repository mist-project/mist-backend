package producer

import (
	"context"
	"fmt"
	"log/slog"
	"mist/src/faults"
	"mist/src/protos/v1/appuser"
	"mist/src/protos/v1/channel"
	"mist/src/protos/v1/event"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

type MessageProducer interface {
	SendMessage(interface{}, event.ActionType, []*appuser.Appuser) error
}

// func (m *MockRedis) Get(ctx context.Context, key string) *MockRedisIntCmd {
// 	args := m.Called(ctx, key)
// 	return args.Get(0).(*MockRedisIntCmd)
// }

// func (m *MockRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *MockRedisIntCmd {
// 	args := m.Called(ctx, key, value, expiration)
// 	return args.Get(0).(*MockRedisIntCmd)
// }

// func (m *MockRedis) Del(ctx context.Context, keys ...string) *MockRedisIntCmd {
// 	args := m.Called(ctx, keys)
// 	return args.Get(0).(*MockRedisIntCmd)
// }

// func (m *MockRedis) Publish(ctx context.Context, channel string, message interface{}) *MockRedisIntCmd {
// 	args := m.Called(channel, message)
// 	return args.Get(0).(*MockRedisIntCmd)
// }

// func (m *MockRedisIntCmd) Result() (int64, error) {
// 	args := m.Called()
// 	return ReturnIfError[int64](args, 1)
// }

type RedisInterface interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd
}

type MProducer struct {
	Redis RedisInterface
}

func NewMProducer(redis RedisInterface) *MProducer {
	return &MProducer{Redis: redis}
}

func (mp *MProducer) SendMessage(
	ctx context.Context, redisChannel string, data interface{}, action event.ActionType, appusers []*appuser.Appuser,
) error {
	msg, err := mp.marshall(data, action, appusers)

	if err != nil {
		return faults.ExtendError(err)
	}

	_, err = mp.Redis.Publish(ctx, redisChannel, msg).Result()

	if err != nil {
		return faults.MessageProducerError(fmt.Sprintf("error sending data to redis: %v", err), slog.LevelError)
	}

	return err
}

func (mp *MProducer) marshall(data interface{}, action event.ActionType, appusers []*appuser.Appuser) ([]byte, error) {
	var e *event.Event

	if appusers == nil {
		appusers = []*appuser.Appuser{}
	}

	switch action {
	case event.ActionType_ACTION_ADD_CHANNEL:
		d, ok := data.(*channel.Channel)

		if !ok {
			return nil, faults.MarshallError(fmt.Sprintf("invalid data for action %v", action), slog.LevelWarn)
		}

		e = &event.Event{
			Meta: &event.Meta{Action: action, Appusers: appusers},
			Data: &event.Event_AddChannel{
				AddChannel: &event.AddChannel{
					Channel: d,
				},
			},
		}
	case event.ActionType_ACTION_LIST_CHANNELS:
		d, ok := data.([]*channel.Channel)
		if !ok {
			return nil, faults.MarshallError(fmt.Sprintf("invalid data for action %v", action), slog.LevelWarn)
		}

		e = &event.Event{
			Meta: &event.Meta{Action: action, Appusers: appusers},
			Data: &event.Event_ListChannels{
				ListChannels: &event.ListChannels{
					Channels: d,
				},
			},
		}
	}

	return proto.Marshal(e)
}
