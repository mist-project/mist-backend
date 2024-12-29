package rpcs

import (
	"testing"

	pb_servers "mist/src/protos/server/v1"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ----- RPC Channels -----
func TestListChannels(t *testing.T) {
	t.Run("returns_nothing_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestClient.ListChannels(
			ctx, &pb_servers.ListChannelsRequest{Name: wrapperspb.String("random")},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetChannels()))
	})

	t.Run("returns_all_resources_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		testChannel(t, nil)
		testChannel(t, nil)

		// ACT
		response, err := TestClient.ListChannels(ctx, &pb_servers.ListChannelsRequest{})
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetChannels()))
	})

	t.Run("can_filter_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		testChannel(t, nil)
		channelToFilterBy := testChannel(t, nil)

		// ACT
		response, err := TestClient.ListChannels(
			ctx, &pb_servers.ListChannelsRequest{AppserverId: wrapperspb.String(channelToFilterBy.AppserverID.String())},
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
		ctx := setup(t, func() {})
		channel := testChannel(t, nil)

		// ACT
		response, err := TestClient.GetByIdChannel(
			ctx, &pb_servers.GetByIdChannelRequest{Id: channel.ID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, channel.ID.String(), response.GetChannel().Id)
	})

	t.Run("invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestClient.GetByIdChannel(
			ctx, &pb_servers.GetByIdChannelRequest{Id: uuid.NewString()},
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
		ctx := setup(t, func() {})

		// ACT
		response, err := TestClient.GetByIdChannel(
			ctx, &pb_servers.GetByIdChannelRequest{Id: "foo"},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.Unknown, s.Code())
		assert.Contains(t, s.Message(), "invalid UUID")
	})
}

// ----- RPC CreateChannel -----
func TestCreateChannel(t *testing.T) {
	t.Run("creates_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		appserver := testAppserver(t, uuid.NewString(), nil)

		// ACT
		response, err := TestClient.CreateChannel(
			ctx, &pb_servers.CreateChannelRequest{Name: "new channel", AppserverId: appserver.ID.String()})
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.NotNil(t, response.Channel)
	})

	t.Run("invalid_arguments_returns_error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestClient.CreateChannel(ctx, &pb_servers.CreateChannelRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "missing name attribute")
	})
}

// ----- RPC DeleteChannel -----
func TestDeleteChannel(t *testing.T) {
	t.Run("deletes_successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		channel := testChannel(t, nil)

		// ACT
		response, err := TestClient.DeleteChannel(ctx, &pb_servers.DeleteChannelRequest{Id: channel.ID.String()})

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("invalid_id_returns_not_found_error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestClient.DeleteChannel(ctx, &pb_servers.DeleteChannelRequest{Id: uuid.NewString()})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "no rows were deleted")
	})
}
