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

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/pressly/goose"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb_mistbe "mist/src/protos/mistbe/v1"
	"mist/src/psql_db/qx"
)

var testServer *grpc.Server
var TestClient pb_mistbe.MistBEServiceClient
var testClientConn *grpc.ClientConn

var dbcPool *pgxpool.Pool
var lis net.Listener

var once sync.Once

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

	testServer = grpc.NewServer()
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

func setup(t *testing.T, ctx context.Context, cleanup func()) {
	t.Cleanup(func() {
		teardown(ctx)
		cleanup()
	})

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
func test_appserver(t *testing.T, appserver *qx.Appserver) *qx.Appserver {
	// Define attributes
	var name string

	if appserver != nil {
		// Custom values
		name = appserver.Name
	} else {
		// Default values
		name = fmt.Sprintf("%s - %s", faker.Word(), uuid.NewString())
	}

	as, err := qx.New(dbcPool).CreateAppserver(context.Background(), name)
	if err != nil {
		t.Fatalf("Unable to create appserver. Error: %v", err)
	}
	return &as
}

func test_channel(t *testing.T, channel *qx.Channel) *qx.Channel {
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
		appserverId = test_appserver(t, nil).ID

	}

	c, err := qx.New(dbcPool).CreateChannel(
		context.Background(), qx.CreateChannelParams{Name: name, AppserverID: appserverId})
	if err != nil {
		t.Fatalf("Unable to create appserver. Error: %v", err)
	}
	return &c
}
