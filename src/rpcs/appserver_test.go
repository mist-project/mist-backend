package rpcs

import (
	"testing"

	pb_mistbe "mist/src/protos/mistbe/v1"
	"mist/src/psql_db/qx"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ----- RPC Appservers -----
func TestListAppServer(t *testing.T) {
	t.Run("can returns nothing successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestClient.ListAppservers(
			ctx, &pb_mistbe.ListAppserversRequest{Name: wrapperspb.String("random")},
		)
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 0, len(response.GetAppservers()))
	})

	t.Run("can return all resources successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		test_appserver(t, nil)
		test_appserver(t, &qx.Appserver{Name: "another one"})

		// ACT
		response, err := TestClient.ListAppservers(ctx, &pb_mistbe.ListAppserversRequest{})
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, 2, len(response.GetAppservers()))
	})

	t.Run("can filter successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
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
		assert.Equal(t, 1, len(response.GetAppservers()))
	})
}

// ----- RPC GetByIdAppserver -----

func TestGetByIdAppServer(t *testing.T) {
	t.Run("returns successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		appserver := test_appserver(t, nil)

		// ACT
		response, err := TestClient.GetByIdAppserver(
			ctx, &pb_mistbe.GetByIdAppserverRequest{Id: appserver.ID.String()},
		)

		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.Equal(t, appserver.ID.String(), response.GetAppserver().Id)
	})

	t.Run("invalid ID returns NotFound error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestClient.GetByIdAppserver(
			ctx, &pb_mistbe.GetByIdAppserverRequest{Id: uuid.NewString()},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "resource not found")
	})

	t.Run("invalid UUID returns parsing error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestClient.GetByIdAppserver(
			ctx, &pb_mistbe.GetByIdAppserverRequest{Id: "foo"},
		)
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.Unknown, s.Code())
		assert.Contains(t, s.Message(), "invalid UUID")
	})
}

// ----- RPC CreateAppserver -----
func TestCreateAppsever(t *testing.T) {
	t.Run("creates successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestClient.CreateAppserver(ctx, &pb_mistbe.CreateAppserverRequest{Name: "someone"})
		if err != nil {
			t.Fatalf("Error performing request %v", err)
		}

		// ASSERT
		assert.NotNil(t, response.Appserver)
	})

	t.Run("invalid arguments returns error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestClient.CreateAppserver(ctx, &pb_mistbe.CreateAppserverRequest{})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, s.Code())
		assert.Contains(t, s.Message(), "missing name attribute")
	})
}

// ----- RPC Deleteappserver -----
func TestDeleteAppserver(t *testing.T) {
	t.Run("deletes successfully", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})
		appserver := test_appserver(t, nil)

		// ACT
		response, err := TestClient.DeleteAppserver(ctx, &pb_mistbe.DeleteAppserverRequest{Id: appserver.ID.String()})

		// ASSERT
		assert.NotNil(t, response)
		assert.Nil(t, err)
	})

	t.Run("invalid ID returns NotFound error", func(t *testing.T) {
		// ARRANGE
		ctx := setup(t, func() {})

		// ACT
		response, err := TestClient.DeleteAppserver(ctx, &pb_mistbe.DeleteAppserverRequest{Id: uuid.NewString()})
		s, ok := status.FromError(err)

		// ASSERT
		assert.Nil(t, response)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, s.Code())
		assert.Contains(t, s.Message(), "no rows were deleted")
	})
}
