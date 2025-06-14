package producer_test

import (
	"bytes"
	"context"
	"mist/src/logging/logger"
	"mist/src/producer"
	"mist/src/protos/v1/channel"
	"mist/src/protos/v1/event"
	"mist/src/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPool_AddJob(t *testing.T) {

	t.Run("Success:job_added_to_queue", func(t *testing.T) {
		// ARRANGE
		wp := producer.NewWorkerPool(2, 10)
		ctx := context.Background()
		mockRedis := new(testutil.MockRedis)
		mockData := &channel.Channel{}
		notification := producer.NewNotificationJob(
			ctx,
			"channel",
			mockData,
			event.ActionType_ACTION_ADD_CHANNEL,
			nil,
			mockRedis,
		)

		// ACT
		wp.AddJob(notification)

		// ASSERT
		assert.Equal(t, wp.GetJobQueueSize(), 1, "Expected job queue size to be 1")
	})

	t.Run("Error:worker_stopped", func(t *testing.T) {
		// ARRANGE
		wp := producer.NewWorkerPool(3, 10)
		wp.StartWorkers()
		wp.Stop() // Cancel the context immediately
		mockRedis := new(testutil.MockRedis)
		mockData := &channel.Channel{}
		notification := producer.NewNotificationJob(
			context.Background(),
			"channel",
			mockData,
			event.ActionType_ACTION_ADD_CHANNEL,
			nil,
			mockRedis,
		)

		// ACT
		wp.AddJob(notification)

		// ASSERT
		assert.Equal(t, wp.GetJobQueueSize(), 0, "Expected job queue size to be 0 due to cancelled context")
	})

	t.Run("Error:when_job_execute_fails_it_logs", func(t *testing.T) {
		// ARRANGE
		var buf bytes.Buffer
		logger.SetLogOutput(&buf) // redirect logger output

		wp := producer.NewWorkerPool(2, 10)
		ctx := context.Background()

		mockRedis := new(testutil.MockRedis)
		mockData := &channel.Channel{}

		notification := producer.NewNotificationJob(
			ctx,
			"channel",
			mockData,
			event.ActionType_ACTION_LIST_CHANNELS, // This should trigger a failure
			nil,
			mockRedis,
		)

		wp.StartWorkers()

		// ACT
		wp.AddJob(notification)

		// Wait for the job to be processed
		wp.Stop()

		// ASSERT
		logOutput := buf.String()
		assert.Contains(t, logOutput, "ERROR", "")
		assert.Contains(t, logOutput, "invalid data for action", "Expected log to contain error message for invalid data")
	})
}

func TestWorkerPool_Stop(t *testing.T) {
	t.Run("Success:worker_pool_stops_and_clears_job_queue", func(t *testing.T) {
		// ARRANGE
		wp := producer.NewWorkerPool(2, 10)
		wp.StartWorkers()

		// ACT
		wp.Stop()

		// ASSERT
		assert.Equal(t, wp.GetJobQueueSize(), 0, "Expected job queue size to be 0 after stopping the worker pool")
	})
}
