package rpcs_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pressly/goose"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"mist/src/middleware"
	pb_appserver "mist/src/protos/v1/appserver"
	pb_appuser "mist/src/protos/v1/appuser"
	pb_channel "mist/src/protos/v1/channel"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
)

var (
	testServer          *grpc.Server
	TestAppserverClient pb_appserver.AppserverServiceClient
	TestAppuserClient   pb_appuser.AppuserServiceClient
	TestChannelClient   pb_channel.ChannelServiceClient
	testClientConn      *grpc.ClientConn

	dbcPool *pgxpool.Pool
	lis     net.Listener

	once sync.Once

	ctxUserKey = "userRequestId"
)

func TestMain(m *testing.M) {
	// ---- SETUP -----
	runTestDbMigrations()
	setupTestAppserverGRPCServiceAndClient()

	// ----- EXECUTION -----
	exit := m.Run()

	// ----- CLEANUP -----
	rpcTestCleanup()
	os.Exit(exit)
}

// ----- SETUP FUNCTION -----

func runTestDbMigrations() {
	// runs test migrations before starting the suite
	once.Do(func() {
		dbConn, err := sql.Open("postgres", os.Getenv("TEST_DATABASE_URL"))

		if err != nil {
			log.Fatalf("Unble to connect to test DB for migrations. %v", err)
		}

		defer dbConn.Close()

		mDir := fmt.Sprintf(
			"%s/%s", os.Getenv("PROJECT_ROOT_PATH"), os.Getenv("GOOSE_MIGRATION_DIR"),
		)

		// Reset DB to starting point ( no migrations )
		if err = goose.Reset(dbConn, mDir); err != nil {
			log.Fatal("Error running migrations: ", err)
		}

		// install all migrations
		if err = goose.Up(dbConn, mDir); err != nil {
			log.Fatal("Error running migrations: ", err)
		}
	})
}

func setupTestAppserverGRPCServiceAndClient() {
	// Creates a grpc server and client to run tests on
	var (
		err error
		lis net.Listener
	)
	dbcPool, err = pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))

	if lis, err = net.Listen("tcp", ":0"); err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	testServer = grpc.NewServer(grpc.ChainUnaryInterceptor(middleware.AuthJwtInterceptor))

	pb_appserver.RegisterAppserverServiceServer(testServer, &rpcs.AppserverGRPCService{DbcPool: dbcPool})
	pb_channel.RegisterChannelServiceServer(testServer, &rpcs.ChannelGRPCService{DbcPool: dbcPool})
	pb_appuser.RegisterAppuserServiceServer(testServer, &rpcs.AppuserGRPCService{DbcPool: dbcPool})

	go func() {
		if err := testServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Setup client connection
	testClientConn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	TestAppserverClient = pb_appserver.NewAppserverServiceClient(testClientConn)
	TestChannelClient = pb_channel.NewChannelServiceClient(testClientConn)
	TestAppuserClient = pb_appuser.NewAppuserServiceClient(testClientConn)

}

func rpcTestCleanup() {
	// Cleans up all the pointers after suite is finished
	if testServer != nil {
		testServer.GracefulStop() // Gracefully shut down the server
	}

	if dbcPool != nil {
		dbcPool.Close()
	}

	if testClientConn != nil {
		testClientConn.Close()
	}
}

func setup(t *testing.T, cleanup func()) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	userRequestId := uuid.NewString()
	ctx = context.WithValue(ctx, ctxUserKey, userRequestId)

	t.Cleanup(func() {
		teardown(ctx)
		cleanup()
		cancel()
	})

	token := createJwtToken(
		t,
		&CreateTokenParams{
			iss:       os.Getenv("MIST_API_JWT_ISSUER"),
			aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
			secretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
			userId:    userRequestId,
		},
	)

	grpcMeta := metadata.Pairs(
		"authorization", fmt.Sprintf("Bearer %s", token),
	)

	ctx = metadata.NewOutgoingContext(ctx, grpcMeta)

	return ctx
}

func teardown(ctx context.Context) {
	// Cleans all the table's data after each test (used in setup) function
	tables := []string{
		"appserver",
		"appuser",
		"channel",
		"appserver_sub",
		"appserver_role",
		"appserver_role_sub"}

	for _, table := range tables {
		query := fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY CASCADE;`, table)

		if _, err := dbcPool.Exec(ctx, query); err != nil {
			log.Fatalf("Failed to truncate table: %v", err)
		}
	}
}

// ----- HELPER FUNCTIONS -----

type CreateTokenParams struct {
	iss       string
	aud       []string
	secretKey string
	userId    string
}

func createJwtToken(t *testing.T, params *CreateTokenParams) string {
	// Define secret key for signing the token

	// Define JWT claims
	c := &middleware.CustomJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   params.iss,
			Audience: params.aud,

			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: params.userId,
	}
	// Create a new token with specified claims
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	// Sign the token using the secret key
	token, err := tok.SignedString([]byte(params.secretKey))
	if err != nil {
		t.Fatalf("error signing the token %v", err)
	}
	return token
}

// ----- DB HELPER FUNCTIONS -----
func testAppuser(t *testing.T, appuser *qx.Appuser) *qx.Appuser {
	var (
		err error
	)
	ctx := context.Background()

	if appuser == nil {
		// Default values
		id, _ := uuid.NewUUID()
		appuser = &qx.Appuser{
			ID:       id,
			Username: faker.Word(),
		}
	}

	user, err := qx.New(dbcPool).CreateAppuser(ctx, qx.CreateAppuserParams{
		ID:       appuser.ID,
		Username: appuser.Username,
	})

	if err != nil {
		t.Fatalf("Unable to create appserver. Error: %v", err)
	}

	return &user
}

func testAppserver(t *testing.T, appserver *qx.Appserver) *qx.Appserver {

	if appserver == nil {
		// Custom values
		appserver = &qx.Appserver{
			AppuserID: testAppuser(t, nil).ID,
			Name:      faker.Word(),
		}
	}

	as, err := qx.New(dbcPool).CreateAppserver(context.Background(), qx.CreateAppserverParams{
		AppuserID: appserver.AppuserID,
		Name:      appserver.Name,
	})

	if err != nil {
		t.Fatalf("Unable to create appserver. Error: %v", err)
	}

	return &as
}

func testAppserverSub(t *testing.T, aSub *qx.AppserverSub) *qx.AppserverSub {
	// Define attributes

	if aSub == nil {
		appuser := testAppuser(t, nil)
		appserver := testAppserver(t, nil)
		aSub = &qx.AppserverSub{
			AppserverID: appserver.ID,
			AppuserID:   appuser.ID,
		}
	}

	asSub, err := qx.New(dbcPool).CreateAppserverSub(
		context.Background(),
		qx.CreateAppserverSubParams{AppserverID: aSub.AppserverID, AppuserID: aSub.AppuserID},
	)

	if err != nil {
		t.Fatalf("Unable to create appserverSub. Error: %v", err)
	}

	return &asSub
}

func testAppserverRole(t *testing.T, aRole *qx.AppserverRole) *qx.AppserverRole {
	// Define attributes

	if aRole == nil {
		aRole = &qx.AppserverRole{
			AppserverID: testAppserver(t, nil).ID,
			Name:        faker.Word(),
		}
	}

	asRole, err := qx.New(dbcPool).CreateAppserverRole(
		context.Background(),
		qx.CreateAppserverRoleParams{AppserverID: aRole.AppserverID, Name: aRole.Name},
	)

	if err != nil {
		t.Fatalf("Unable to create appserverRole. Error: %v", err)
	}

	return &asRole
}

func testAppserverRoleSub(t *testing.T, roleSub *qx.AppserverRoleSub) *qx.AppserverRoleSub {
	// Define attributes

	if roleSub == nil {
		// Custom values
		user := testAppuser(t, nil)
		appserver := testAppserver(t, nil)
		sub := testAppserverSub(t, &qx.AppserverSub{
			AppserverID: appserver.ID,
			AppuserID:   user.ID,
		})
		role := testAppserverRole(
			t,
			&qx.AppserverRole{Name: "some random role", AppserverID: appserver.ID},
		)
		roleSub = &qx.AppserverRoleSub{
			AppserverRoleID: role.ID,
			AppserverSubID:  sub.ID,
			AppserverID:     appserver.ID,
			AppuserID:       user.ID,
		}
	}

	asrSub, err := qx.New(dbcPool).CreateAppserverRoleSub(
		context.Background(),
		qx.CreateAppserverRoleSubParams{
			AppserverRoleID: roleSub.AppserverRoleID,
			AppserverSubID:  roleSub.AppserverSubID,
			AppuserID:       roleSub.AppuserID,
			AppserverID:     roleSub.AppserverID,
		},
	)

	if err != nil {
		t.Fatalf("Unable to create appserverRole. Error: %v", err)
	}

	return &asrSub
}

func testChannel(t *testing.T, c *qx.Channel) *qx.Channel {
	// Define attributes

	if c == nil {
		// Default value
		c = &qx.Channel{
			Name:        faker.Word(),
			AppserverID: testAppserver(t, nil).ID,
		}
	}

	channel, err := qx.New(dbcPool).CreateChannel(
		context.Background(), qx.CreateChannelParams{Name: c.Name, AppserverID: c.AppserverID})

	if err != nil {
		t.Fatalf("Unable to create appserver. Error: %v", err)
	}
	return &channel
}
