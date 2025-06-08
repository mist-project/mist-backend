package mist_redis_test

import (
	"mist/src/producer/mist_redis"
	"os"
	"os/exec"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// This is the helper process function
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// We call the real function — expecting it to fatal
	mist_redis.ConnectToRedis(os.Getenv("REDIS_DB"))
}

func TestConnectToRedis_Success(t *testing.T) {
	// ARRANGE
	os.Setenv("REDIS_USERNAME", "default")
	os.Setenv("REDIS_PASSWORD", "yourpassword")
	os.Setenv("REDIS_HOSTNAME", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_NOTIFICATION_CHANNEL", "test-channel")

	dbStr := "0"

	defer func() {
		// Cleanup
		os.Unsetenv("REDIS_USERNAME")
		os.Unsetenv("REDIS_PASSWORD")
		os.Unsetenv("REDIS_HOSTNAME")
		os.Unsetenv("REDIS_PORT")
		os.Unsetenv("REDIS_NOTIFICATION_CHANNEL")
	}()

	// ACT
	client := mist_redis.ConnectToRedis(dbStr)

	// ASSERT
	assert.NotNil(t, client)
	assert.IsType(t, &redis.Client{}, client)
}

// Helper to run a test in a subprocess and check if it exits (for testing log.Fatal)
func TestConnectToRedis_MissingNotificationChannel(t *testing.T) {
	// ARRANGE
	os.Setenv("REDIS_USERNAME", "default")
	os.Setenv("REDIS_PASSWORD", "yourpassword")
	os.Setenv("REDIS_HOSTNAME", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_NOTIFICATION_CHANNEL", "") // <--- Missing on purpose
	dbStr := "0"

	// ACT
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess")
	cmd.Env = append(os.Environ(),
		"GO_WANT_HELPER_PROCESS=1",
		"REDIS_DB="+dbStr,
	)

	output, err := cmd.CombinedOutput()

	// ASSERT
	// We expect the process to exit with non-zero
	if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() != 0 {
		// Success — fatal happened
		assert.Contains(t, string(output), "REDIS_NOTIFICATION_CHANNEL environment variable is not set")
	} else {
		t.Fatalf("process ran successfully when we expected fatal. Output: %s", string(output))
	}
}

func TestConnectToRedis_MissingRedisHost(t *testing.T) {
	// ARRANGE
	os.Setenv("REDIS_USERNAME", "default")
	os.Setenv("REDIS_PASSWORD", "yourpassword")
	os.Setenv("REDIS_HOSTNAME", "") // <--- Missing on purpose
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_NOTIFICATION_CHANNEL", "test-channel")
	dbStr := "0"

	// ACT
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess")
	cmd.Env = append(os.Environ(),
		"GO_WANT_HELPER_PROCESS=1",
		"REDIS_DB="+dbStr,
	)

	output, err := cmd.CombinedOutput()

	// ASSERT
	if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() != 0 {
		assert.Contains(t, string(output), "Redis connection details are not set in environment variables")
	} else {
		t.Fatalf("process ran successfully when we expected fatal. Output: %s", string(output))
	}
}

func TestConnectToRedis_Invalid_Db(t *testing.T) {
	// ARRANGE
	os.Setenv("REDIS_USERNAME", "default")
	os.Setenv("REDIS_PASSWORD", "yourpassword")
	os.Setenv("REDIS_HOSTNAME", "ok")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_NOTIFICATION_CHANNEL", "test-channel")
	dbStr := "invalid"

	// ACT
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess")
	cmd.Env = append(os.Environ(),
		"GO_WANT_HELPER_PROCESS=1",
		"REDIS_DB="+dbStr,
	)

	output, err := cmd.CombinedOutput()

	// ASSERT
	if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() != 0 {
		assert.Contains(t, string(output), "Invalid REDIS_DB value")
	} else {
		t.Fatalf("process ran successfully when we expected fatal. Output: %s", string(output))
	}
}
