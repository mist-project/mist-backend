package service_test

import (
	"fmt"
	"mist/src/psql_db/qx"
	"mist/src/service"
	"mist/src/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListChannels(t *testing.T) {
	t.Run("error_is_returned_when_creating_server_fails", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		expectedRequest := qx.CreateAppserverParams{Name: "foo", AppuserID: parsedUid}
		mockTxQuerier := new(testutil.MockQuerier)
		mockTxQuerier.On("CreateAppserver", mock.Anything, expectedRequest).Return(qx.Appserver{}, fmt.Errorf("a db error"))

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("WithTx", mock.Anything).Return(mockTxQuerier)

		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier)

		// // ACT
		_, err := svc.Create(expectedRequest)

		// // ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "(-3): database error- ")
	})

	t.Run("error_is_returned_when_creating_appserver_sub_fails", func(t *testing.T) {
		// ARRANGE
		ctx := testutil.Setup(t, func() {})
		parsedUid, _ := uuid.Parse(ctx.Value(testutil.CtxUserKey).(string))
		expectedRequest := qx.CreateAppserverParams{Name: "foo", AppuserID: parsedUid}
		mockTxQuerier := new(testutil.MockQuerier)
		mockTxQuerier.On("CreateAppserver", mock.Anything, expectedRequest).Return(
			qx.Appserver{ID: uuid.New()}, nil,
		)
		mockTxQuerier.On("CreateAppserver", mock.Anything, mock.Anything).Return(qx.Appserver{})

		mockQuerier := new(testutil.MockQuerier)
		mockQuerier.On("WithTx", mock.Anything).Return(mockTxQuerier)

		svc := service.NewAppserverService(ctx, testutil.TestDbConn, mockQuerier)

		// // ACT
		_, err := svc.Create(expectedRequest)

		// // ASSERT
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "(-3): database error- ")
	})

}
