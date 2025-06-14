package testutil

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

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pressly/goose"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"mist/src/middleware"
	"mist/src/producer"
	"mist/src/protos/v1/appserver"
	"mist/src/protos/v1/appserver_role"
	"mist/src/protos/v1/appserver_role_sub"
	"mist/src/protos/v1/appserver_sub"
	"mist/src/protos/v1/appuser"
	"mist/src/protos/v1/channel"
	"mist/src/protos/v1/channel_role"
	"mist/src/psql_db/db"
	"mist/src/rpcs"
)

var (
	testServer                 *grpc.Server
	TestAppserverClient        appserver.AppserverServiceClient
	TestAppserverRoleClient    appserver_role.AppserverRoleServiceClient
	TestAppserverRoleSubClient appserver_role_sub.AppserverRoleSubServiceClient
	TestAppserverSubClient     appserver_sub.AppserverSubServiceClient
	TestAppuserClient          appuser.AppuserServiceClient
	TestChannelClient          channel.ChannelServiceClient
	TestChannelRoleClient      channel_role.ChannelRoleServiceClient
	testClientConn             *grpc.ClientConn

	TestDbConn        *pgxpool.Pool
	TestKProducer     = new(MockProducer)
	mockRedis         = new(MockRedis)
	MockRedisProducer = producer.NewMProducer(mockRedis)
	TestMockAuth      = new(MockAuthorizer)

	once sync.Once

	CtxUserKey    = "userRequestId"
	DefaultUserId = "571637fd-3c1e-4bb5-9077-e35edbe02526"
)

func ReturnIfError[T any](args mock.Arguments, index int) (T, error) {
	err := args.Error(index)
	var zero T
	if err != nil {
		return zero, err
	}
	return args.Get(0).(T), nil
}

// ----- SETUP FUNCTION -----

func SetupDbMigrations() {
	// runs test migrations before starting the suite
	TestDbConn, err := sql.Open("postgres", os.Getenv("TEST_DATABASE_URL"))

	if err != nil {
		log.Fatalf("Unble to connect to test DB for migrations. %v", err)
	}

	defer TestDbConn.Close()

	mDir := fmt.Sprintf(
		"%s/%s", os.Getenv("PROJECT_ROOT_PATH"), os.Getenv("GOOSE_MIGRATION_DIR"),
	)

	// Reset DB to starting point ( no migrations )
	if err = goose.Reset(TestDbConn, mDir); err != nil {
		log.Fatal("Error running migrations: ", err)
	}

	// install all migrations
	if err = goose.Up(TestDbConn, mDir); err != nil {
		log.Fatal("Error running migrations: ", err)
	}

}

func SetupDbConnection() {
	var err error

	once.Do(func() {
		TestDbConn, err = pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
		if err != nil {
			log.Fatalf("error runing on dbconnection %v", err)
		}
	})
}

func SetupTestGRPCServicesAndClient() {
	// Creates a grpc server and client to run tests on
	var (
		err error
		lis net.Listener
	)

	// Base MockAuth to skip authorization in tests
	TestMockAuth.On("Authorize", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	if lis, err = net.Listen("tcp", ":0"); err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	interceptors, _ := rpcs.BaseInterceptors()
	testServer = grpc.NewServer(interceptors)

	// for now we will mock all the producer calls to be successful. unit tests should
	// ensure that the producer is called where it should happen

	TestKProducer.On("SendMessage", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockRedis.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(redis.NewIntCmd(context.Background()))
	rpcs.RegisterGrpcServices(testServer, &rpcs.GrpcDependencies{
		Db:        db.NewQuerier(TestDbConn),
		MProducer: MockRedisProducer,
	})

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

	TestAppuserClient = appuser.NewAppuserServiceClient(testClientConn)
	TestAppserverClient = appserver.NewAppserverServiceClient(testClientConn)
	TestAppserverRoleClient = appserver_role.NewAppserverRoleServiceClient(testClientConn)
	TestAppserverRoleSubClient = appserver_role_sub.NewAppserverRoleSubServiceClient(testClientConn)
	TestAppserverSubClient = appserver_sub.NewAppserverSubServiceClient(testClientConn)
	TestChannelClient = channel.NewChannelServiceClient(testClientConn)
	TestChannelRoleClient = channel_role.NewChannelRoleServiceClient(testClientConn)
}

func RpcTestCleanup() {
	// Cleans up all the pointers after suite is finished
	if testServer != nil {
		testServer.GracefulStop() // Gracefully shut down the server
	}

	if TestDbConn != nil {
		TestDbConn.Close()
	}

	if testClientConn != nil {
		testClientConn.Close()
	}
}

func Setup(t *testing.T, cleanup func()) (context.Context, db.Querier) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	DefaultUserId = uuid.NewString()
	ctx = context.WithValue(ctx, CtxUserKey, DefaultUserId)

	q, err := db.NewQuerier(TestDbConn).Begin(ctx)

	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	t.Cleanup(func() {
		q.Rollback(ctx)
		cleanup()
		cancel()
	})

	token, claims := CreateJwtToken(
		t,
		&CreateTokenParams{
			Iss:       os.Getenv("MIST_API_JWT_ISSUER"),
			Aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
			SecretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
			UserId:    DefaultUserId,
		},
	)

	grpcMeta := metadata.Pairs(
		"authorization", fmt.Sprintf("Bearer %s", token),
	)

	ctx = metadata.NewOutgoingContext(ctx, grpcMeta)
	ctx = context.WithValue(ctx, middleware.JwtClaimsK, claims)

	return ctx, q
}

// ----- HELPER FUNCTIONS -----

type CreateTokenParams struct {
	Iss       string
	Aud       []string
	SecretKey string
	UserId    string
}

func CreateJwtToken(t *testing.T, params *CreateTokenParams) (string, *middleware.CustomJWTClaims) {
	// Define secret key for signing the token

	// Define JWT claims
	c := &middleware.CustomJWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   params.Iss,
			Audience: params.Aud,

			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: params.UserId,
	}
	// Create a new token with specified claims
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, c)

	// Sign the token using the secret key
	token, err := tok.SignedString([]byte(params.SecretKey))
	if err != nil {
		t.Fatalf("error signing the token %v", err)
	}
	return token, c
}
