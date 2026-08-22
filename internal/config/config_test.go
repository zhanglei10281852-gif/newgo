package config

import (
	"os"
	"testing"
)

func TestDefaultsLoad(t *testing.T) {
	for _, key := range []string{"HTTP_ADDR", "DATABASE_PATH", "BUSINESS_TIMEZONE", "SESSION_TTL", "PERMIT_TTL", "WORKER_INTERVAL", "WORKER_BATCH_SIZE", "SHUTDOWN_TIMEOUT"} {
		_ = os.Unsetenv(key)
	}
	c, e := Load()
	if e != nil {
		t.Fatal(e)
	}
	if c.HTTPAddr == "" || c.DatabasePath == "" || c.WorkerBatchSize < 1 {
		t.Fatalf("config=%+v", c)
	}
}
func TestInvalidDuration(t *testing.T) {
	t.Setenv("SESSION_TTL", "not-duration")
	if _, e := Load(); e == nil {
		t.Fatal("invalid duration accepted")
	}
}
func TestInvalidTimezone(t *testing.T) {
	t.Setenv("BUSINESS_TIMEZONE", "Mars/Phobos")
	if _, e := Load(); e == nil {
		t.Fatal("invalid timezone accepted")
	}
}
