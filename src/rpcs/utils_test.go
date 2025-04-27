package rpcs_test

import (
	"os"
	"testing"

	"mist/src/testutil"
)

func TestMain(m *testing.M) {
	// ---- SETUP -----
	testutil.SetupDbConnection()
	testutil.SetupTestGRPCServicesAndClient()

	// ----- EXECUTION -----
	exit := m.Run()

	// ----- CLEANUP -----
	testutil.RpcTestCleanup()
	os.Exit(exit)
}
