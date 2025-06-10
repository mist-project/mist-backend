package service_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"mist/src/faults"
	"mist/src/faults/message"
	"mist/src/producer"
	"mist/src/protos/v1/appserver"
	"mist/src/protos/v1/appuser"
	"mist/src/psql_db/qx"
	"mist/src/service"
	"mist/src/testutil"
)

func TestAppserverService_PgTypeToPb(t *testing.T) {

	// ARANGE
	ctx := context.Background()
	svc := service.NewAppserverService(
		ctx,
		&service.ServiceDeps{
			Db:        new(testutil.MockQuerier),
			MProducer: producer.NewMProducer(new(testutil.MockRedis)),
		},
	)

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

	t.Run("Success:creation_on_valid_ops", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})

		server := &qx.Appserver{ID: uuid.New(), Name: "test-app", AppuserID: uuid.New()}
		expectedRequest := qx.CreateAppserverParams{Name: server.Name, AppuserID: server.AppuserID}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("CreateAppserver", mock.Anything, expectedRequest).Return(
			qx.Appserver{ID: server.ID, Name: server.Name}, nil,
		)
		mockQuerier.On("CreateAppserverSub", mock.Anything, mock.Anything).Return(qx.AppserverSub{}, nil)

		// Service initialization
		svc := service.NewAppserverService(ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer})

		// ACT
		response, err := svc.Create(expectedRequest)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, server.ID, response.ID)
		mockQuerier.AssertExpectations(t)
	})

	// t.Run("Error:is_returned_when_starting_tx_fails", func(t *testing.T) {
	// 	// ARRANGE
	// 	badConnection, err := pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
	// 	badConnection.Close()

	// 	if err != nil {
	// 		t.Fatalf("failed to start db connection")
	// 	}

	// 	ctx, _ := testutil.Setup(t, func() {})
	// 	parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
	// 	expectedRequest := qx.CreateAppserverParams{Name: "foo", AppuserID: parsedUid}

	// 	mockQuerier := new(testutil.MockQuerier)
	// 	mockRedis := new(testutil.MockRedis)
	// 	producer := producer.NewMProducer(mockRedis)

	// 	svc := service.NewAppserverService(
	// 		ctx, &service.ServiceDeps{DbConn: badConnection, Db: mockQuerier, MProducer: producer},
	// 	)

	// 	// ACT
	// 	_, err = svc.Create(expectedRequest)

	// 	// ASSERT
	// 	assert.NotNil(t, err)
	// 	assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
	// 	testutil.AssertCustomErrorContains(t, err, "tx initialization error: closed pool")
	// 	mockQuerier.AssertExpectations(t)
	// })

	t.Run("Error:is_returned_when_creating_server_fails", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		expectedRequest := qx.CreateAppserverParams{Name: "foo", AppuserID: uuid.New()}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("CreateAppserver", mock.Anything, expectedRequest).Return(nil, fmt.Errorf("a db error"))

		svc := service.NewAppserverService(ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer})

		// // ACT
		_, err := svc.Create(expectedRequest)

		// // ASSERT
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "create appserver error: a db error")
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Error:is_returned_when_creating_appserver_sub_fails", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		server := qx.Appserver{ID: uuid.New(), Name: "test-app", AppuserID: uuid.New()}
		expectedRequest := qx.CreateAppserverParams{Name: "foo", AppuserID: server.AppuserID}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("CreateAppserver", mock.Anything, expectedRequest).Return(server, nil)
		mockQuerier.On("CreateAppserverSub", mock.Anything, mock.Anything).Return(
			nil, fmt.Errorf("a db error"),
		)

		svc := service.NewAppserverService(ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer})

		// // ACT
		_, err := svc.Create(expectedRequest)

		// // ASSERT
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "create appserver sub error: a db error")
		mockQuerier.AssertExpectations(t)
	})
}

func TestAppserverService_GetById(t *testing.T) {

	t.Run("Success:appserver_return", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverId := uuid.New()
		expected := qx.Appserver{ID: appserverId, Name: "test-app"}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetAppserverById", ctx, appserverId).Return(expected, nil)

		svc := service.NewAppserverService(ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer})

		// ACT
		actual, err := svc.GetById(appserverId)

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected.ID, actual.ID)
		assert.Equal(t, expected.Name, actual.Name)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Error:returns_not_found_when_no_rows", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverId := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetAppserverById", ctx, appserverId).
			Return(nil, fmt.Errorf(message.DbNotFound))

		svc := service.NewAppserverService(ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer})

		// ACT
		_, err := svc.GetById(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.NotFoundMessage)
		testutil.AssertCustomErrorContains(t, err, "unable to find appserver with id")
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Error:returns_database_error_on_failure", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		appserverId := uuid.New()

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("GetAppserverById", ctx, appserverId).
			Return(nil, fmt.Errorf("connection reset by peer"))

		svc := service.NewAppserverService(ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer})

		// ACT
		_, err := svc.GetById(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: connection reset by peer")
		mockQuerier.AssertExpectations(t)
	})
}

func TestAppserverService_List(t *testing.T) {

	t.Run("Success:with_name_filter", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		ownerID := uuid.New()
		nameFilter := "test-app"
		expected := []qx.Appserver{{ID: uuid.New(), Name: nameFilter, AppuserID: ownerID}}
		name := pgtype.Text{Valid: true, String: nameFilter}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("ListAppservers", ctx, mock.MatchedBy(func(p qx.ListAppserversParams) bool {
			return p.AppuserID == ownerID && p.Name.Valid && p.Name.String == nameFilter
		})).Return(expected, nil)

		svc := service.NewAppserverService(ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer})

		// ACT
		result, err := svc.List(qx.ListAppserversParams{Name: name, AppuserID: ownerID})

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Success:without_name_filter", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		ownerID := uuid.New()
		expected := []qx.Appserver{{ID: uuid.New(), Name: "app-1", AppuserID: ownerID}}
		name := pgtype.Text{Valid: false, String: ""}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("ListAppservers", ctx, mock.MatchedBy(func(p qx.ListAppserversParams) bool {
			return p.AppuserID == ownerID && !p.Name.Valid
		})).Return(expected, nil)

		svc := service.NewAppserverService(ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer})

		// ACT
		result, err := svc.List(qx.ListAppserversParams{Name: name, AppuserID: ownerID})

		// ASSERT
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
		mockQuerier.AssertExpectations(t)
	})

	t.Run("Error:failure_on_db_error", func(t *testing.T) {
		// ARRANGE
		ctx, _ := testutil.Setup(t, func() {})
		ownerID := uuid.New()
		name := pgtype.Text{Valid: false, String: ""}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("ListAppservers", ctx, mock.Anything).
			Return([]qx.Appserver(nil), fmt.Errorf("some db error"))

		svc := service.NewAppserverService(ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer})

		// ACT
		_, err := svc.List(qx.ListAppserversParams{Name: name, AppuserID: ownerID})

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: some db error")
		mockQuerier.AssertExpectations(t)
	})
}

func TestAppserverService_Delete(t *testing.T) {

	ctx, _ := testutil.Setup(t, func() {})
	appserverId := uuid.New()

	t.Run("Success:deletion", func(t *testing.T) {
		// ARRANGE

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("DeleteAppserver", ctx, appserverId).Return(int64(1), nil)
		mockQuerier.On("ListAppserverUserSubs", ctx, appserverId).Return([]qx.ListAppserverUserSubsRow{}, nil)

		svc := service.NewAppserverService(ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer})

		// ACT
		err := svc.Delete(appserverId)

		// ASSERT
		assert.NoError(t, err)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Success:deletion_with_subs_sends_notification", func(t *testing.T) {
		// ARRANGE
		subs := []qx.ListAppserverUserSubsRow{
			{AppuserID: uuid.New(), AppuserUsername: "user1"},
			{AppuserID: uuid.New(), AppuserUsername: "user2"},
		}
		users := make([]*appuser.Appuser, 0, len(subs))
		for _, sub := range subs {
			users = append(users, &appuser.Appuser{
				Id:       sub.AppuserID.String(),
				Username: sub.AppuserUsername,
			})
		}

		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("DeleteAppserver", ctx, appserverId).Return(int64(1), nil)
		mockQuerier.On("ListAppserverUserSubs", ctx, appserverId).Return(subs, nil)
		mockRedis.On("Publish", ctx, os.Getenv("REDIS_NOTIFICATION_CHANNEL"), mock.Anything).Return(redis.NewIntCmd(ctx))

		svc := service.NewAppserverService(ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer})

		_ = svc.Delete(appserverId)

		// ACT
		err := svc.Delete(appserverId)

		// ASSERT
		assert.NoError(t, err)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:on_no_rows_deleted", func(t *testing.T) {
		// ARRANGE
		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("ListAppserverUserSubs", ctx, appserverId).Return([]qx.ListAppserverUserSubsRow{}, nil)
		mockQuerier.On("DeleteAppserver", ctx, appserverId).Return(int64(0), nil)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		svc := service.NewAppserverService(ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer})

		// ACT
		err := svc.Delete(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.NotFoundMessage)
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:on_db_list_subs_failure", func(t *testing.T) {
		// ARRANGE
		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("ListAppserverUserSubs", ctx, appserverId).Return(nil, fmt.Errorf("db failure"))

		svc := service.NewAppserverService(ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer})

		// ACT
		err := svc.Delete(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db failure")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	t.Run("Error:on_db_delete_failure", func(t *testing.T) {
		// ARRANGE
		mockQuerier := new(testutil.MockQuerier)
		mockRedis := new(testutil.MockRedis)
		producer := producer.NewMProducer(mockRedis)

		mockQuerier.On("ListAppserverUserSubs", ctx, appserverId).Return([]qx.ListAppserverUserSubsRow{}, nil)
		mockQuerier.On("DeleteAppserver", ctx, appserverId).Return(nil, fmt.Errorf("db failure"))

		svc := service.NewAppserverService(ctx, &service.ServiceDeps{Db: mockQuerier, MProducer: producer})

		// ACT
		err := svc.Delete(appserverId)

		// ASSERT
		assert.Error(t, err)
		assert.Equal(t, err.Error(), faults.DatabaseErrorMessage)
		testutil.AssertCustomErrorContains(t, err, "database error: db failure")
		mockQuerier.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})
}
