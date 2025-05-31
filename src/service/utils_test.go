package service_test

import (
	"mist/src/logging/logger"
	"mist/src/testutil"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// ---- SETUP -----
	logger.InitializeLogger()
	testutil.SetupDbConnection()

	// ----- EXECUTION -----
	exit := m.Run()

	// ----- CLEANUP -----
	os.Exit(exit)
}
