package service_test

import (
	"mist/src/testutil"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// ---- SETUP -----
	testutil.SetupDbConnection()

	// ----- EXECUTION -----
	exit := m.Run()

	// ----- CLEANUP -----
	os.Exit(exit)
}
