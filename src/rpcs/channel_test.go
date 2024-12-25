package rpcs

import (
	"context"
	"fmt"
	"testing"
	"time"

	pb_mistbe "mist/src/protos/mistbe/v1"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ----- RPC Channels -----
func TestChannelsReturnsNothingSuccessfully(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})

	// ACT
	response, err := TestClient.ListChannels(
		ctx, &pb_mistbe.ListChannelsRequest{Name: wrapperspb.String("random")},
	)
	if err != nil {
		t.Fatalf("Error performing request %v", err)
	}

	// ASSERT
	assert.Equal(t, 0, len(response.GetChannels()))
}

func TestChannelsReturnsAllResourcesSuccessfully(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})
	test_channel(t, nil)
	test_channel(t, nil)

	// ACT
	response, err := TestClient.ListChannels(ctx, &pb_mistbe.ListChannelsRequest{})
	if err != nil {
		t.Fatalf("Error performing request %v", err)
	}

	// ASSERT
	assert.Equal(t, 2, len(response.GetChannels()))
}

func TestChannelsCanFilterSuccessfully(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})
	test_channel(t, nil)
	channelToFilterBy := test_channel(t, nil)

	// ACT
	response, err := TestClient.ListChannels(
		ctx, &pb_mistbe.ListChannelsRequest{Name: wrapperspb.String(channelToFilterBy.Name)},
	)
	if err != nil {
		t.Fatalf("Error performing request %v", err)
	}

	// ASSERT
	fmt.Printf("\nresponse: %v", response.GetChannels())
	assert.Equal(t, 1, len(response.GetChannels()))
}

// ----- RPC GetByIdChannel -----
func TestGetByIdChannelReturnsSuccessfully(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})
	channel := test_channel(t, nil)

	// ACT
	response, err := TestClient.GetByIdChannel(
		ctx, &pb_mistbe.GetByIdChannelRequest{Id: channel.ID.String()},
	)

	if err != nil {
		t.Fatalf("Error performing request %v", err)
	}

	// ASSERT
	assert.Equal(t, channel.ID.String(), response.GetChannel().Id)
}

func TestGetByIdChannelInvalidIdReturnsNotFoundError(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})

	// ACT
	response, err := TestClient.GetByIdChannel(
		ctx, &pb_mistbe.GetByIdChannelRequest{Id: uuid.NewString()},
	)
	s, ok := status.FromError(err)

	// ASSERT
	assert.Nil(t, response)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, s.Code())
	assert.Contains(t, s.Message(), "resource not found")
}

func TestGetByIdChannelInvalidUuidReturnsParsingError(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})

	// ACT
	response, err := TestClient.GetByIdChannel(
		ctx, &pb_mistbe.GetByIdChannelRequest{Id: "foo"},
	)
	s, ok := status.FromError(err)

	// ASSERT
	assert.Nil(t, response)
	assert.True(t, ok)
	assert.Equal(t, codes.Unknown, s.Code())
	assert.Contains(t, s.Message(), "invalid UUID")
}

// ----- RPC CreateChannel -----
func TestCreateChannelSuccessfully(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})
	appserver := test_appserver(t, nil)

	// ACT
	response, err := TestClient.CreateChannel(
		ctx, &pb_mistbe.CreateChannelRequest{Name: "new channel", AppserverId: appserver.ID.String()})
	if err != nil {
		t.Fatalf("Error performing request %v", err)
	}

	// ASSERT
	assert.NotNil(t, response.Channel)
}

func TestCreateChannelInvalidArgsError(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})

	// ACT
	response, err := TestClient.CreateChannel(ctx, &pb_mistbe.CreateChannelRequest{})
	s, ok := status.FromError(err)

	// ASSERT
	assert.Nil(t, response)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, s.Code())
	assert.Contains(t, s.Message(), "missing name attribute")
}

// ----- RPC DeleteChannel -----
func TestDeleteChannelSuccessfully(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})
	channel := test_channel(t, nil)

	// ACT
	response, err := TestClient.DeleteChannel(ctx, &pb_mistbe.DeleteChannelRequest{Id: channel.ID.String()})

	// ASSERT
	assert.NotNil(t, response)
	assert.Nil(t, err)
}

func TestDeleteChannelInvalidIdNotFoundError(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})

	// ACT
	response, err := TestClient.DeleteChannel(ctx, &pb_mistbe.DeleteChannelRequest{Id: uuid.NewString()})
	s, ok := status.FromError(err)

	// ASSERT
	assert.Nil(t, response)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, s.Code())
	assert.Contains(t, s.Message(), "no rows were deleted")
}
