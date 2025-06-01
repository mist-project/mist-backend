package rpcs_test

import (
	"os"
	"testing"

	"mist/src/logging/logger"
	"mist/src/testutil"
)

func TestMain(m *testing.M) {
	// ---- SETUP -----
	logger.InitializeLogger()
	testutil.SetupDbConnection()
	testutil.SetupTestGRPCServicesAndClient()

	// ----- EXECUTION -----
	exit := m.Run()

	// ----- CLEANUP -----
	testutil.RpcTestCleanup()
	os.Exit(exit)
}
