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
	pb_channel "mist/src/protos/channel/v1"
	pb_server "mist/src/protos/server/v1"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
)

var testServer *grpc.Server
var TestAppserverClient pb_server.ServerServiceClient
var TestChannelClient pb_channel.ChannelServiceClient
var testClientConn *grpc.ClientConn

var dbcPool *pgxpool.Pool
var lis net.Listener

var once sync.Once

var ctxUserKey = "userRequestId"

func TestMain(m *testing.M) {
	// ---- SETUP -----
	runTestDbMigrations()
	setupTestAppserverGRPCServiceAndClient()

	// ----- EXECUTION -----
	exitValue := m.Run()

	// ----- CLEANUP -----
	rpcTestCleanup()
	os.Exit(exitValue)
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

		migrationsDir := fmt.Sprintf("%s/%s", os.Getenv("PROJECT_ROOT_PATH"), os.Getenv("GOOSE_MIGRATION_DIR"))
		// Reset DB to starting point ( no migrations )
		err = goose.Reset(dbConn, migrationsDir)
		if err != nil {
			log.Fatal("Error running migrations: ", err)
		}

		// install all migrations
		err = goose.Up(dbConn, migrationsDir)
		if err != nil {
			log.Fatal("Error running migrations: ", err)
		}
	})
}

func setupTestAppserverGRPCServiceAndClient() {
	// Creates a grpc server and client to run tests on
	var err error
	dbcPool, err = pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	testServer = grpc.NewServer(grpc.ChainUnaryInterceptor(middleware.AuthJwtInterceptor))

	pb_server.RegisterServerServiceServer(testServer, &rpcs.AppserverGRPCService{DbcPool: dbcPool})
	pb_channel.RegisterChannelServiceServer(testServer, &rpcs.ChannelGRPCService{DbcPool: dbcPool})

	go func() {
		if err := testServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Setup client connection
	testClientConn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	TestAppserverClient = pb_server.NewServerServiceClient(testClientConn)
	TestChannelClient = pb_channel.NewChannelServiceClient(testClientConn)

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
	tokenStr := createJwtToken(
		t,
		&CreateTokenParams{
			iss:       os.Getenv("MIST_API_JWT_ISSUER"),
			aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
			secretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
			userId:    userRequestId,
		},
	)

	grpcMetadata := metadata.Pairs(
		"authorization", fmt.Sprintf("Bearer %s", tokenStr),
	)

	ctx = metadata.NewOutgoingContext(ctx, grpcMetadata)

	return ctx
}

func teardown(ctx context.Context) {
	// Cleans all the table's data after each test (used in setup) function
	tables := []string{"appserver", "channel", "appserver_sub", "appserver_role"}
	for _, table := range tables {
		query := fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY CASCADE;`, table)
		_, err := dbcPool.Exec(ctx, query)
		if err != nil {
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
	claims := &middleware.CustomJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   params.iss,
			Audience: params.aud,

			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: params.userId,
	}
	// Create a new token with specified claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token using the secret key
	tokenString, err := token.SignedString([]byte(params.secretKey))
	if err != nil {
		t.Fatalf("error signing the token %v", err)
	}
	return tokenString
}

// ----- DB HELPER FUNCTIONS -----
func testAppserver(t *testing.T, userId string, appserver *qx.Appserver) *qx.Appserver {
	// Define attributes
	var name string
	parsedUserId, err := uuid.Parse(userId)

	if err != nil {
		t.Fatalf("unable to create appserver. Error %v", err)
	}

	if appserver != nil {
		// Custom values
		name = appserver.Name

	} else {
		// Default values
		name = faker.Word()
	}

	as, err := qx.New(dbcPool).CreateAppserver(context.Background(), qx.CreateAppserverParams{
		Name:    name,
		OwnerID: parsedUserId,
	})
	if err != nil {
		t.Fatalf("Unable to create appserver. Error: %v", err)
	}
	return &as
}

func testAppserverSub(t *testing.T, userId string, appserverSub *qx.AppserverSub) *qx.AppserverSub {
	// Define attributes
	var appserverId uuid.UUID
	ownerId, err := uuid.Parse(userId)

	if err != nil {
		t.Fatalf("unable to create appserverSub. Error %v", err)
	}

	if appserverSub != nil {
		// Custom values
		appserverId = appserverSub.AppserverID
	} else {
		appserverId = testAppserver(t, userId, nil).ID
	}

	asSub, err := qx.New(dbcPool).CreateAppserverSub(context.Background(), qx.CreateAppserverSubParams{
		AppserverID: appserverId, OwnerID: ownerId,
	})
	if err != nil {
		t.Fatalf("Unable to create appserverSub. Error: %v", err)
	}
	return &asSub
}

func testAppserverRole(t *testing.T, userId string, appserverRole *qx.AppserverRole) *qx.AppserverRole {
	// Define attributes
	var appserverId uuid.UUID
	var name string
	if appserverRole != nil {
		// Custom values
		appserverId = appserverRole.AppserverID
		name = appserverRole.Name
	} else {
		appserverId = testAppserver(t, userId, nil).ID
		name = faker.Word()
	}

	asRole, err := qx.New(dbcPool).CreateAppserverRole(context.Background(), qx.CreateAppserverRoleParams{
		AppserverID: appserverId, Name: name,
	})

	if err != nil {
		t.Fatalf("Unable to create appserverRole. Error: %v", err)
	}
	return &asRole
}

func testAppserverRoleSub(t *testing.T, userId string, appserverRoleSub *qx.AppserverRoleSub) *qx.AppserverRoleSub {
	// Define attributes
	var appserverRoleId uuid.UUID
	var appserverSubId uuid.UUID

	ownerId, err := uuid.Parse(userId)

	if err != nil {
		t.Fatalf("unable to create appserverSub. Error %v", err)
	}

	if appserverRoleSub != nil {
		// Custom values
		appserverRoleId = appserverRoleSub.AppserverRoleID
		appserverSubId = appserverRoleSub.AppserverSubID
	} else {
		appserverRole := testAppserverRole(t, userId, nil)
		appserverRoleId = appserverRole.ID
		appserverSubId = testAppserverSub(
			t, userId, &qx.AppserverSub{AppserverID: appserverRole.AppserverID, OwnerID: ownerId},
		).ID
	}

	asrSub, err := qx.New(dbcPool).CreateAppserverRoleSub(context.Background(), qx.CreateAppserverRoleSubParams{
		AppserverRoleID: appserverRoleId, AppserverSubID: appserverSubId,
	})

	if err != nil {
		t.Fatalf("Unable to create appserverRole. Error: %v", err)
	}
	return &asrSub
}

func testChannel(t *testing.T, channel *qx.Channel) *qx.Channel {
	// Define attributes
	var appserverId uuid.UUID
	var name string

	if channel != nil {
		// Custom values
		name = channel.Name
		appserverId = channel.ID
	} else {
		// Default values
		name = faker.Word()
		appserverId = testAppserver(t, uuid.NewString(), nil).ID

	}

	c, err := qx.New(dbcPool).CreateChannel(
		context.Background(), qx.CreateChannelParams{Name: name, AppserverID: appserverId})
	if err != nil {
		t.Fatalf("Unable to create appserver. Error: %v", err)
	}
	return &c
}
