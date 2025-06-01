package service_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/faults"
	"mist/src/faults/message"
	"mist/src/protos/v1/appserver"
	"mist/src/protos/v1/appuser"
	"mist/src/protos/v1/event"
	"mist/src/psql_db/qx"
	"mist/src/service"
	"mist/src/testutil"
)

func TestAppserverService_PgTypeToPb(t *testing.T) {

	// ARANGE
	ctx := context.Background()
	svc := service.NewAppserverService(ctx, testutil.TestDbConn, new(testutil.MockQuerier), new(testutil.MockProducer))

	id := uuid.New()
	now := time.Now()

	server := &qx.Appserver{
		ID:   id,
		Name: "example",
		CreatedAt: pgtype.Timestamp{
			Time:  now,
			Valid: true,
		},
	}

	expected := &appserver.Appserver{
		Id:        id.String(),
		Name:      "example",
		CreatedAt: timestamppb.New(now),
	}

	// ACT
	result := svc.PgTypeToPb(server)

	// ASSERT
	assert.Equal(t, expected, result)
}

func TestAppserverService_Create(t *testing.T) {

	t.Run("Successful:creation_on_valid_ops", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))

		testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "user bar"}, false)
		appserver := testutil.TestAppserver(t, &qx.Appserver{Name: "foo", AppuserID: parsedUid}, false)

		expectedRequest := qx.CreateAppserverParams{Name: appserver.Name, AppuserID: parsedUid}

		mockTxQuerier := new(testutil.MockQuerier)
		mockTxQuerier.On("CreateAppserver", mock.Anything, expectedRequest).Return(
			qx.Appserver{ID: appserver.ID, Name: appserver.Name}, nil,
		)
		mockTxQuerier.On("CreateAppserverSub", mock.Anything, mock.Anything).Return(
			qx.AppserverSub{}, nil,
		)
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("WithTx", mock.Anything).Return(mockTxQuerier)

		// Service initialization
		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		response, err := svc.Create(expectedRequest)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, appserver.ID, response.ID)
	})

	t.Run("Error:is_returned_when_starting_tx_fails", func(t *testing.T) {
		// ARRANGE
		badConnection, err := pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
		badConnection.Close()

		if err != nil {
			t.Fatalf("failed to start db connection")
		}

		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		expectedRequest := qx.CreateAppserverParams{Name: "foo", AppuserID: parsedUid}
		mockQuerier := new(testutil.MockQuerier)

		svc := service.NewAppserverService(ctx, badConnection, mockQuerier, new(testutil.MockProducer))

		// // ACT
		_, err = svc.Create(expectedRequest)

		// // ASSERT
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "tx initialization error: closed pool")
	})

	t.Run("Error:is_returned_when_creating_server_fails", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		expectedRequest := qx.CreateAppserverParams{Name: "foo", AppuserID: parsedUid}
		mockTxQuerier := new(testutil.MockQuerier)
		mockTxQuerier.On("CreateAppserver", mock.Anything, expectedRequest).Return(nil, fmt.Errorf("a db error"))

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("WithTx", mock.Anything).Return(mockTxQuerier)

		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// // ACT
		_, err := svc.Create(expectedRequest)

		// // ASSERT
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "create appserver error: a db error")
	})

	t.Run("Error:is_returned_when_creating_appserver_sub_fails", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		expectedRequest := qx.CreateAppserverParams{Name: "foo", AppuserID: parsedUid}
		mockTxQuerier := new(testutil.MockQuerier)
		mockTxQuerier.On("CreateAppserver", mock.Anything, expectedRequest).Return(qx.Appserver{}, nil)
		mockTxQuerier.On("CreateAppserverSub", mock.Anything, mock.Anything).Return(
			qx.AppserverSub{}, fmt.Errorf("a db error"),
		)

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("WithTx", mock.Anything).Return(mockTxQuerier)

		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// // ACT
		_, err := svc.Create(expectedRequest)

		// // ASSERT
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "create appserver sub error: a db error")
	})

	t.Run("Error:commit_fails_with_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		testutil.TestAppuser(t, &qx.Appuser{ID: parsedUid, Username: "user bar"}, false)
		appserver := testutil.TestAppserver(t, &qx.Appserver{Name: "foo", AppuserID: parsedUid}, false)

		expectedServer := qx.CreateAppserverParams{Name: appserver.Name, AppuserID: parsedUid}
		expectedSub := qx.CreateAppserverSubParams{AppserverID: appserver.ID, AppuserID: parsedUid}

		mockTx := new(testutil.MockTx)
		mockTx.On("Commit", mock.Anything).Return(fmt.Errorf("commit failed"))
		mockTxQuerier := new(testutil.MockQuerier)

		mockTxQuerier.On("CreateAppserver", mock.Anything, expectedServer).Return(
			qx.Appserver{ID: appserver.ID, Name: appserver.Name}, nil,
		)
		mockTxQuerier.On("CreateAppserverSub", mock.Anything, expectedSub).Return(
			qx.AppserverSub{}, nil,
		)

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("WithTx", mock.Anything).Return(mockTxQuerier)

		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		_, err := svc.CreateWithTx(expectedServer, mockTx)

		// ASSERT
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error commit: commit failed")
		mockTx.AssertExpectations(t)
	})
}

func TestAppserverService_GetById(t *testing.T) {

	t.Run("Successful:appserver_return", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		expected := qx.Appserver{ID: appserverId, Name: "test-app"}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverById", ctx, appserverId).Return(expected, nil)

		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		actual, err := svc.GetById(appserverId)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, actual.ID)
		assert.Equal(t, expected.Name, actual.Name)
	})

	t.Run("Error:returns_not_found_when_no_rows", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverById", ctx, appserverId).
			Return(nil, fmt.Errorf(message.DbNotFound))

		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		_, err := svc.GetById(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.NotFoundMessage)
		testutil.AssertCustomErrorContains(t, err, "unable to find appserver with id")
	})

	t.Run("Error:returns_database_error_on_failure", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("GetAppserverById", ctx, appserverId).
			Return(nil, fmt.Errorf("connection reset by peer"))

		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))

		// ACT
		_, err := svc.GetById(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: connection reset by peer")
	})
}

func TestAppserverService_List(t *testing.T) {

	t.Run("Successful:with_name_filter", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		ownerID := uuid.New()
		nameFilter := "test-app"
		expected := []qx.Appserver{
			{ID: uuid.New(), Name: nameFilter, AppuserID: ownerID},
		}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListAppservers", ctx, mock.MatchedBy(func(p qx.ListAppserversParams) bool {
			return p.AppuserID == ownerID && p.Name.Valid && p.Name.String == nameFilter
		})).Return(expected, nil)

		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))
		var name = pgtype.Text{Valid: true, String: nameFilter}

		// ACT
		result, err := svc.List(qx.ListAppserversParams{Name: name, AppuserID: ownerID})

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Successful:without_name_filter", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		ownerID := uuid.New()
		expected := []qx.Appserver{
			{ID: uuid.New(), Name: "app-1", AppuserID: ownerID},
		}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListAppservers", ctx, mock.MatchedBy(func(p qx.ListAppserversParams) bool {
			return p.AppuserID == ownerID && !p.Name.Valid
		})).Return(expected, nil)

		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))
		var name = pgtype.Text{Valid: false, String: ""}

		// ACT
		result, err := svc.List(qx.ListAppserversParams{Name: name, AppuserID: ownerID})

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Error:failure_on_db_error", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		ownerID := uuid.New()
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListAppservers", ctx, mock.Anything).
			Return([]qx.Appserver(nil), fmt.Errorf("some db error"))

		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier, new(testutil.MockProducer))
		var name = pgtype.Text{Valid: false, String: ""}

		// ACT
		_, err := svc.List(qx.ListAppserversParams{Name: name, AppuserID: ownerID})

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: some db error")
	})
}

func TestAppserverService_Delete(t *testing.T) {

	ctx := testutil.Setup(t, func() {})
	appserverId := uuid.New()

	t.Run("Successful:deletion", func(t *testing.T) {
		// ARRANGE\
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserver", ctx, appserverId).Return(int64(1), nil)
		mockQuerier.On("ListAppserverUserSubs", ctx, appserverId).Return([]qx.ListAppserverUserSubsRow{}, nil)

		mockProducer := new(testutil.MockProducer)
		mockProducer.On("SendMessage", mock.Anything, event.ActionType_ACTION_REMOVE_SERVER, []*appuser.Appuser{}).Return(nil)

		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		err := svc.Delete(appserverId)

		// ASSERT
		assert.NoError(t, err)
	})

	t.Run("Successful:deletion_with_subs_sends_notification", func(t *testing.T) {
		// ARRANGE
		subs := []qx.ListAppserverUserSubsRow{
			{ID: uuid.New(), Username: "user1"},
			{ID: uuid.New(), Username: "user2"},
		}
		users := make([]*appuser.Appuser, 0, len(subs))
		for _, sub := range subs {
			users = append(users, &appuser.Appuser{
				Id:       sub.ID.String(),
				Username: sub.Username,
			})
		}

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("DeleteAppserver", ctx, appserverId).Return(int64(1), nil)
		mockQuerier.On("ListAppserverUserSubs", ctx, appserverId).Return(subs, nil)

		mockProducer := new(testutil.MockProducer)
		mockProducer.On("SendMessage", mock.Anything, event.ActionType_ACTION_REMOVE_SERVER, users).Return(nil)

		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)
		svc.Delete(appserverId)

		// ACT
		err := svc.Delete(appserverId)

		// ASSERT
		assert.NoError(t, err)
		mockProducer.AssertExpectations(t)

	})

	t.Run("Error:on_no_rows_deleted", func(t *testing.T) {
		// ARRANGE
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListAppserverUserSubs", ctx, appserverId).Return([]qx.ListAppserverUserSubsRow{}, nil)
		mockQuerier.On("DeleteAppserver", ctx, appserverId).Return(int64(0), nil)
		mockProducer := new(testutil.MockProducer)

		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		err := svc.Delete(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.NotFoundMessage)
	})

	t.Run("Error:on_db_list_subs_failure", func(t *testing.T) {
		// ARRANGE
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListAppserverUserSubs", ctx, appserverId).Return(nil, fmt.Errorf("db failure"))
		mockProducer := new(testutil.MockProducer)

		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		err := svc.Delete(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db failure")
	})

	t.Run("Error:on_db_delete_failure", func(t *testing.T) {
		// ARRANGE
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListAppserverUserSubs", ctx, appserverId).Return([]qx.ListAppserverUserSubsRow{}, nil)
		mockQuerier.On("DeleteAppserver", ctx, appserverId).Return(nil, fmt.Errorf("db failure"))
		mockProducer := new(testutil.MockProducer)

		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier, mockProducer)

		// ACT
		err := svc.Delete(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db failure")
	})
}
