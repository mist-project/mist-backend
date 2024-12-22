package rpcs

import (
	"context"
	"testing"
	"time"

	pb_mistbe "mist/src/protos/mistbe/v1"
	"mist/src/psql_db/qx"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ----- RPC Appservers -----
func TestAppserversReturnsNothingSuccessfully(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})

	// ACT
	response, err := TestClient.ListAppservers(
		ctx, &pb_mistbe.ListAppserversRequest{Name: wrapperspb.String("random")},
	)
	if err != nil {
		t.Fatalf("Error performing request %v", err)
	}

	// ASSERT
	assert.Equal(t, len(response.GetAppservers()), 0)
}

func TestAppserversReturnsAllResourcesSuccessfully(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})
	test_appserver(t, nil)
	test_appserver(t, &qx.Appserver{Name: "another one"})

	// ACT
	response, err := TestClient.ListAppservers(ctx, &pb_mistbe.ListAppserversRequest{})
	if err != nil {
		t.Fatalf("Error performing request %v", err)
	}

	// ASSERT
	assert.Equal(t, len(response.GetAppservers()), 2)
}

func TestAppserversCanFilterSuccessfully(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})
	test_appserver(t, nil)
	test_appserver(t, &qx.Appserver{Name: "another one"})

	// ACT
	response, err := TestClient.ListAppservers(
		ctx, &pb_mistbe.ListAppserversRequest{Name: wrapperspb.String("another one")},
	)
	if err != nil {
		t.Fatalf("Error performing request %v", err)
	}

	// ASSERT
	assert.Equal(t, len(response.GetAppservers()), 1)
}

// ----- RPC GetByIdAppserver -----
func TestGetByIdAppserversReturnsSuccessfully(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})
	appserver := test_appserver(t, nil)

	// ACT
	response, err := TestClient.GetByIdAppserver(
		ctx, &pb_mistbe.GetByIdAppserverRequest{Id: appserver.ID.String()},
	)

	if err != nil {
		t.Fatalf("Error performing request %v", err)
	}

	// ASSERT
	assert.Equal(t, response.GetAppserver().Id, appserver.ID.String())
}

func TestGetByIdAppserversInvalidIdReturnsNotFoundError(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})

	// ACT
	response, err := TestClient.GetByIdAppserver(
		ctx, &pb_mistbe.GetByIdAppserverRequest{Id: uuid.NewString()},
	)
	s, ok := status.FromError(err)

	// ASSERT
	assert.Nil(t, response)
	assert.True(t, ok)
	assert.Equal(t, s.Code(), codes.NotFound)
	assert.Contains(t, s.Message(), "resource not found")
}

// ----- RPC CreateAppserver -----
func TestCreateAppserverSuccessfully(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})

	// ACT
	response, err := TestClient.CreateAppserver(ctx, &pb_mistbe.CreateAppserverRequest{Name: "someone"})
	if err != nil {
		t.Fatalf("Error performing request %v", err)
	}

	// ASSERT
	assert.NotNil(t, response.Appserver)
}

func TestCreateAppserverInvalidArgsError(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})

	// ACT
	response, err := TestClient.CreateAppserver(ctx, &pb_mistbe.CreateAppserverRequest{})
	s, ok := status.FromError(err)

	// ASSERT
	assert.Nil(t, response)
	assert.True(t, ok)
	assert.Equal(t, s.Code(), codes.InvalidArgument)
	assert.Contains(t, s.Message(), "missing name attribute")
}

// ----- RPC Deleteappserver -----
func TestDeleteAppserverSuccessfully(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})
	appserver := test_appserver(t, nil)

	// ACT
	response, err := TestClient.DeleteAppserver(ctx, &pb_mistbe.DeleteAppserverRequest{Id: appserver.ID.String()})

	// ASSERT
	assert.NotNil(t, response)
	assert.Nil(t, err)
}

func TestDeleteAppserverInvalidIdNotFoundError(t *testing.T) {
	// ARRANGE
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	setup(t, ctx, func() {
		cancel()
	})

	// ACT
	response, err := TestClient.DeleteAppserver(ctx, &pb_mistbe.DeleteAppserverRequest{Id: uuid.NewString()})
	s, ok := status.FromError(err)

	// ASSERT
	assert.Nil(t, response)
	assert.True(t, ok)
	assert.Equal(t, s.Code(), codes.NotFound)
	assert.Contains(t, s.Message(), "no rows were deleted")
}
