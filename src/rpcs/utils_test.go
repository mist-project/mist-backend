package rpcs

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
	pb_mistbe "mist/src/protos/mistbe/v1"
	"mist/src/psql_db/qx"
)

var testServer *grpc.Server
var TestClient pb_mistbe.MistBEServiceClient
var testClientConn *grpc.ClientConn

var dbcPool *pgxpool.Pool
var lis net.Listener

var once sync.Once

var ctxUserKey = "userRequestId"

func TestMain(m *testing.M) {
	// ---- SETUP -----
	runTestDbMigrations()
	setupTestGrpcserverAndClient()

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

func setupTestGrpcserverAndClient() {
	// Creates a grpc server and client to run tests on
	var err error
	dbcPool, err = pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	testServer = grpc.NewServer(grpc.ChainUnaryInterceptor(middleware.AuthJwtInterceptor))

	pb_mistbe.RegisterMistBEServiceServer(testServer, &Grpcserver{DbcPool: dbcPool})

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
	TestClient = pb_mistbe.NewMistBEServiceClient(testClientConn)
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
	tokenStr := CreateJWTToken(
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
	tables := []string{"appserver", "channel"}
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

func CreateJWTToken(t *testing.T, params *CreateTokenParams) string {
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
		name = fmt.Sprintf("%s - %s", faker.Word(), uuid.NewString())
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
		name = fmt.Sprintf("%s - %s", faker.Word(), uuid.NewString())
		appserverId = testAppserver(t, uuid.NewString(), nil).ID

	}

	c, err := qx.New(dbcPool).CreateChannel(
		context.Background(), qx.CreateChannelParams{Name: name, AppserverID: appserverId})
	if err != nil {
		t.Fatalf("Unable to create appserver. Error: %v", err)
	}
	return &c
}
