package producer

import (
	"context"
	"fmt"
	"log/slog"
	"mist/src/faults"
	"mist/src/protos/v1/appuser"
	"mist/src/protos/v1/channel"
	"mist/src/protos/v1/event"

	"google.golang.org/protobuf/proto"
)

type Job interface {
	Execute(int) error
	Ctx() context.Context
}

// ------ NOTIFICATION JOB -----
type NotificationJob struct {
	ctx          context.Context
	redisChannel string
	data         interface{}
	action       event.ActionType
	appusers     []*appuser.Appuser
	redisClient  RedisInterface
}

func NewNotificationJob(
	ctx context.Context,
	redisChannel string,
	data interface{},
	action event.ActionType,
	appusers []*appuser.Appuser,
	redisClient RedisInterface,
) *NotificationJob {
	return &NotificationJob{
		ctx:          ctx,
		redisChannel: redisChannel,
		data:         data,
		action:       action,
		appusers:     appusers,
		redisClient:  redisClient,
	}
}

func (job *NotificationJob) Ctx() context.Context {
	return job.ctx
}

func (job *NotificationJob) Execute(worker int) error {
	msg, err := job.marshall(job.data, job.action, job.appusers)

	if err != nil {
		return faults.ExtendError(err)
	}

	_, err = job.redisClient.Publish(job.ctx, job.redisChannel, msg).Result()

	if err != nil {
		return faults.MessageProducerError(fmt.Sprintf("WORKER[%d] error sending data to redis: %v", worker, err), slog.LevelError)
	}

	return err
}

func (job *NotificationJob) marshall(data interface{}, action event.ActionType, appusers []*appuser.Appuser) ([]byte, error) {
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

// ------ NOTIFICATION JOB -----
type StopWorkerJob struct {
	ctx context.Context
}

func NewStopWorkerJob(ctx context.Context) *StopWorkerJob {
	return &StopWorkerJob{
		ctx: ctx,
	}
}
func (job *StopWorkerJob) Ctx() context.Context {
	return job.ctx
}

func (job *StopWorkerJob) Execute(worker int) error {
	return nil // No error, just a signal to stop the worker
}
