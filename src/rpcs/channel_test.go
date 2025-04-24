package rpcs_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	pb_channel "mist/src/protos/v1/channel"
	"mist/src/psql_db/qx"
	"mist/src/testutil"
)

// ----- RPC Channels -----
func TestListChannels(t *testing.T) {
	t.Run("returns_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestChannelClient.ListChannels(
			ctx, &pb_channel.ListChannelsRequest{Name: wrapperspb.String("random")},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetChannels()))
	})

	t.Run("returns_all_resources_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		server := testutil.TestAppserver(t, nil)
		testutil.TestChannel(t, &qx.Channel{Name: "foo", AppserverID: server.ID})
		testutil.TestChannel(t, &qx.Channel{Name: "bar", AppserverID: server.ID})

		// ACT
		response, err := testutil.TestChannelClient.ListChannels(ctx, &pb_channel.ListChannelsRequest{})
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetChannels()))
	})

	t.Run("can_filter_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		server := testutil.TestAppserver(t, nil)
		testutil.TestChannel(t, &qx.Channel{Name: "bar", AppserverID: server.ID})
		testutil.TestChannel(t, nil)

		// ACT
		response, err := testutil.TestChannelClient.ListChannels(
			ctx, &pb_channel.ListChannelsRequest{AppserverId: wrapperspb.String(server.ID.String())},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 1, len(response.GetChannels()))
	})
}

// ----- RPC GetByIdChannel -----
func TestGetByIdChannel(t *testing.T) {
	t.Run("returns_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		channel := testutil.TestChannel(t, nil)

		// ACT
		response, err := testutil.TestChannelClient.GetByIdChannel(
			ctx, &pb_channel.GetByIdChannelRequest{Id: channel.ID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, channel.ID.String(), response.GetChannel().Id)
	})

	t.Run("invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestChannelClient.GetByIdChannel(
			ctx, &pb_channel.GetByIdChannelRequest{Id: uuid.NewString()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "resource not found")
	})

	t.Run("invalid_uuid_returns_parsing_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestChannelClient.GetByIdChannel(
			ctx, &pb_channel.GetByIdChannelRequest{Id: "foo"},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error:\n - id: value must be a valid UUID")
	})
}

// ----- RPC CreateChannel -----
func TestCreateChannel(t *testing.T) {
	t.Run("creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserver := testutil.TestAppserver(t, nil)

		// ACT
		response, err := testutil.TestChannelClient.CreateChannel(
			ctx, &pb_channel.CreateChannelRequest{Name: "new channel", AppserverId: appserver.ID.String()})
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.NotNil(t, response.Channel)
	})

	t.Run("invalid_arguments_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestChannelClient.CreateChannel(ctx, &pb_channel.CreateChannelRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "validation error")
	})
}

// ----- RPC DeleteChannel -----
func TestDeleteChannel(t *testing.T) {
	t.Run("deletes_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		channel := testutil.TestChannel(t, nil)

		// ACT
		response, err := testutil.TestChannelClient.DeleteChannel(ctx, &pb_channel.DeleteChannelRequest{Id: channel.ID.String()})

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})

		// ACT
		response, err := testutil.TestChannelClient.DeleteChannel(ctx, &pb_channel.DeleteChannelRequest{Id: uuid.NewString()})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "no rows were deleted")
	})
}
