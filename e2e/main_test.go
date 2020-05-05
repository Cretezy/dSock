package dsock_test

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Wait for API/worker to be started (usually not needed)
	// time.Sleep(time.Second)

	os.Exit(m.Run())
}
