package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb_appserver "mist/src/protos/v1/appserver"
	"mist/src/psql_db/db"
	"mist/src/psql_db/qx"
)

type AppserverService struct {
	ctx    context.Context
	dbConn *pgxpool.Pool
	db     db.Querier
}

// Creates a new AppserverService struct.
func NewAppserverService(ctx context.Context, dbConn *pgxpool.Pool, db db.Querier) *AppserverService {
	return &AppserverService{ctx: ctx, dbConn: dbConn, db: db}
}

// Converts a database appserver object to protobuff appserver object
func (s *AppserverService) PgTypeToPb(a *qx.Appserver) *pb_appserver.Appserver {
	return &pb_appserver.Appserver{
		Id:        a.ID.String(),
		Name:      a.Name,
		CreatedAt: timestamppb.New(a.CreatedAt.Time),
	}
}

// Creates an appserver, uses CreateWithTx helper function to wrap creation in transaction
// Note: the transaction will be committed in CreateWithTx. The creator of the server gets automatically assigned
// an appserver sub.
func (s *AppserverService) Create(obj qx.CreateAppserverParams) (*qx.Appserver, error) {
	tx, err := s.dbConn.BeginTx(s.ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("(%d) tx initialization error: %v", DatabaseError, err)
	}
	defer tx.Rollback(s.ctx)

	response, err := s.CreateWithTx(obj, tx)

	return response, err
}

// Creates an appserver with provided transaction. This function will commit the transaction.
func (s *AppserverService) CreateWithTx(obj qx.CreateAppserverParams, tx pgx.Tx) (*qx.Appserver, error) {
	txQ := s.db.WithTx(tx)

	appserver, err := txQ.CreateAppserver(s.ctx, obj)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("(%d) create appserver error: %v", DatabaseError, err))
	}

	// once the appserver is created, add user as a subscriber
	_, err = NewAppserverSubService(s.ctx, s.dbConn, s.db).CreateWithTx(
		qx.CreateAppserverSubParams{AppserverID: appserver.ID, AppuserID: obj.AppuserID},
		tx,
	)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("(%d) create appserver sub error: %v", DatabaseError, err))
	}

	if err := tx.Commit(s.ctx); err != nil {
		return nil, fmt.Errorf("(%d) commit - %v", DatabaseError, err)
	}

	return &appserver, err
}

// Gets an appserver detail by its id.
func (s *AppserverService) GetById(id uuid.UUID) (*qx.Appserver, error) {
	appserver, err := s.db.GetAppserverById(s.ctx, id)

	if err != nil {
		// TODO: this check must be a standard db error result checker
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, fmt.Errorf(fmt.Sprintf("(%d) resource not found", NotFoundError))
		}

		return nil, fmt.Errorf(fmt.Sprintf("(%d) database error: %v", DatabaseError, err))
	}

	return &appserver, nil
}

// Lists all appservers based on the owner. Name filter is also added but it may get deprecated.
func (s *AppserverService) List(params qx.ListUserAppserversParams) ([]qx.Appserver, error) {
	appservers, err := s.db.ListUserAppservers(s.ctx, params)

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("(%d) database error: %v", DatabaseError, err))
	}

	return appservers, nil
}

// Delete appserver object, for now only owners can delete an appserver.
func (s *AppserverService) Delete(obj qx.DeleteAppserverParams) error {
	deleted, err := s.db.DeleteAppserver(s.ctx, obj)

	if err != nil {
		return fmt.Errorf(fmt.Sprintf("(%d) database error: %v", DatabaseError, err))
	} else if deleted == 0 {
		return fmt.Errorf(fmt.Sprintf("(%d) no rows were deleted", NotFoundError))
	}

	return err
}
