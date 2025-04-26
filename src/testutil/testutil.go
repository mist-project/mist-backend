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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"mist/src/middleware"
	pb_appserver "mist/src/protos/v1/appserver"
	pb_appserverrole "mist/src/protos/v1/appserver_role"
	pb_appserverrolesub "mist/src/protos/v1/appserver_role_sub"
	pb_appserversub "mist/src/protos/v1/appserver_sub"
	pb_appuser "mist/src/protos/v1/appuser"
	pb_channel "mist/src/protos/v1/channel"
	"mist/src/psql_db/qx"
	"mist/src/rpcs"
)

var (
	testServer                 *grpc.Server
	TestAppserverClient        pb_appserver.AppserverServiceClient
	TestAppserverRoleClient    pb_appserverrole.AppserverRoleServiceClient
	TestAppserverRoleSubClient pb_appserverrolesub.AppserverRoleSubServiceClient
	TestAppserverSubClient     pb_appserversub.AppserverSubServiceClient
	TestAppuserClient          pb_appuser.AppuserServiceClient
	TestChannelClient          pb_channel.ChannelServiceClient
	testClientConn             *grpc.ClientConn

	TestDbConn *pgxpool.Pool
	lis        net.Listener

	once sync.Once

	CtxUserKey = "userRequestId"
)

// ----- SETUP FUNCTION -----

func RunTestDbMigrations() {
	var err error
	// runs test migrations before starting the suite
	once.Do(func() {
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
	})

	TestDbConn, err = pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		log.Fatalf("error runing on dbconnection %v", err)
	}
}

func SetupTestGRPCServicesAndClient() {
	// Creates a grpc server and client to run tests on
	var (
		err error
		lis net.Listener
	)

	if lis, err = net.Listen("tcp", ":0"); err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	interceptors, err := rpcs.BaseInterceptors()

	testServer = grpc.NewServer(interceptors)
	rpcs.RegisterGrpcServices(testServer, TestDbConn)

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

	TestAppuserClient = pb_appuser.NewAppuserServiceClient(testClientConn)
	TestAppserverClient = pb_appserver.NewAppserverServiceClient(testClientConn)
	TestAppserverRoleClient = pb_appserverrole.NewAppserverRoleServiceClient(testClientConn)
	TestAppserverRoleSubClient = pb_appserverrolesub.NewAppserverRoleSubServiceClient(testClientConn)
	TestAppserverSubClient = pb_appserversub.NewAppserverSubServiceClient(testClientConn)
	TestChannelClient = pb_channel.NewChannelServiceClient(testClientConn)
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

func Setup(t *testing.T, cleanup func()) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	userRequestId := uuid.NewString()
	ctx = context.WithValue(ctx, CtxUserKey, userRequestId)

	t.Cleanup(func() {
		teardown(ctx)
		cleanup()
		cancel()
	})

	token, claims := CreateJwtToken(
		t,
		&CreateTokenParams{
			Iss:       os.Getenv("MIST_API_JWT_ISSUER"),
			Aud:       []string{os.Getenv("MIST_API_JWT_AUDIENCE")},
			SecretKey: os.Getenv("MIST_API_JWT_SECRET_KEY"),
			UserId:    userRequestId,
		},
	)

	grpcMeta := metadata.Pairs(
		"authorization", fmt.Sprintf("Bearer %s", token),
	)

	ctx = metadata.NewOutgoingContext(ctx, grpcMeta)
	ctx = context.WithValue(ctx, middleware.JwtClaimsK, claims)
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
		"appserver_role_sub",
	}

	for _, table := range tables {
		query := fmt.Sprintf(`TRUNCATE TABLE %s RESTART IDENTITY CASCADE;`, table)

		if _, err := TestDbConn.Exec(ctx, query); err != nil {
			log.Fatalf("Failed to truncate table: %v", err)
		}
	}
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

// ----- DB HELPER FUNCTIONS -----
func TestAppuser(t *testing.T, appuser *qx.Appuser) *qx.Appuser {
	var (
		err error
	)
	ctx := context.Background()

	if appuser == nil {
		// Default values
		id, _ := uuid.NewUUID()
		appuser = &qx.Appuser{
			ID:       id,
			Username: uuid.NewString(),
		}
	}

	user, err := qx.New(TestDbConn).CreateAppuser(ctx, qx.CreateAppuserParams{
		ID:       appuser.ID,
		Username: appuser.Username,
	})

	if err != nil {
		t.Fatalf("Unable to create appserver. Error: %v", err)
	}

	return &user
}

func TestAppserver(t *testing.T, appserver *qx.Appserver) *qx.Appserver {

	if appserver == nil {
		// Custom values
		appserver = &qx.Appserver{
			AppuserID: TestAppuser(t, nil).ID,
			Name:      uuid.NewString(),
		}
	}

	as, err := qx.New(TestDbConn).CreateAppserver(context.Background(), qx.CreateAppserverParams{
		AppuserID: appserver.AppuserID,
		Name:      appserver.Name,
	})

	if err != nil {
		t.Fatalf("Unable to create appserver. Error: %v", err)
	}

	return &as
}

func TestAppserverSub(t *testing.T, aSub *qx.AppserverSub) *qx.AppserverSub {
	// Define attributes

	if aSub == nil {
		appuser := TestAppuser(t, nil)
		appserver := TestAppserver(t, nil)
		aSub = &qx.AppserverSub{
			AppserverID: appserver.ID,
			AppuserID:   appuser.ID,
		}
	}

	asSub, err := qx.New(TestDbConn).CreateAppserverSub(
		context.Background(),
		qx.CreateAppserverSubParams{AppserverID: aSub.AppserverID, AppuserID: aSub.AppuserID},
	)

	if err != nil {
		t.Fatalf("Unable to create appserverSub. Error: %v", err)
	}

	return &asSub
}

func TestAppserverRole(t *testing.T, aRole *qx.AppserverRole) *qx.AppserverRole {
	// Define attributes

	if aRole == nil {
		aRole = &qx.AppserverRole{
			AppserverID: TestAppserver(t, nil).ID,
			Name:        uuid.NewString(),
		}
	}

	asRole, err := qx.New(TestDbConn).CreateAppserverRole(
		context.Background(),
		qx.CreateAppserverRoleParams{AppserverID: aRole.AppserverID, Name: aRole.Name},
	)

	if err != nil {
		t.Fatalf("Unable to create appserverRole. Error: %v", err)
	}

	return &asRole
}

func TestAppserverRoleSub(t *testing.T, roleSub *qx.AppserverRoleSub) *qx.AppserverRoleSub {
	// Define attributes

	if roleSub == nil {
		// Custom values
		user := TestAppuser(t, nil)
		appserver := TestAppserver(t, nil)
		sub := TestAppserverSub(t, &qx.AppserverSub{
			AppserverID: appserver.ID,
			AppuserID:   user.ID,
		})
		role := TestAppserverRole(
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

	asrSub, err := qx.New(TestDbConn).CreateAppserverRoleSub(
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

func TestChannel(t *testing.T, c *qx.Channel) *qx.Channel {
	// Define attributes

	if c == nil {
		// Default value
		c = &qx.Channel{
			Name:        uuid.NewString(),
			AppserverID: TestAppserver(t, nil).ID,
		}
	}

	channel, err := qx.New(TestDbConn).CreateChannel(
		context.Background(), qx.CreateChannelParams{Name: c.Name, AppserverID: c.AppserverID})

	if err != nil {
		t.Fatalf("Unable to create appserver. Error: %v", err)
	}
	return &channel
}
