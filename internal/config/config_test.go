package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeEnvFile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	return path
}

func TestLoadDefaults(t *testing.T) {
	got, err := Load(writeEnvFile(t, ""))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	wantNats := NatsConfig{Connstr: defaultNatsConnstr, Stream: defaultNatsStream, Consumer: defaultNatsConsumer}
	if got.Nats != wantNats {
		t.Errorf("Nats = %+v, want %+v", got.Nats, wantNats)
	}
	wantClickhouse := ClickhouseConfig{Addr: defaultClickhouseAddr, Database: defaultClickhouseDB, Username: defaultClickhouseUser}
	if got.Clickhouse != wantClickhouse {
		t.Errorf("Clickhouse = %+v, want %+v", got.Clickhouse, wantClickhouse)
	}
}

func TestLoadFromEnvFile(t *testing.T) {
	got, err := Load(writeEnvFile(t, "NATSCONNSTR=nats://example:4222\nNATSSTREAM=events\nNATSCONSUMER=worker\nCLICKHOUSEADDR=clickhouse:9000\nCLICKHOUSEDB=analytics\nCLICKHOUSEUSER=reader\nCLICKHOUSEPASSWORD=secret\n"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got.Nats != (NatsConfig{Connstr: "nats://example:4222", Stream: "events", Consumer: "worker"}) {
		t.Errorf("Nats = %+v", got.Nats)
	}
	if got.Clickhouse != (ClickhouseConfig{Addr: "clickhouse:9000", Database: "analytics", Username: "reader", Password: "secret"}) {
		t.Errorf("Clickhouse = %+v", got.Clickhouse)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.env"))
	if err == nil {
		t.Fatal("Load() error = nil, want an error")
	}
}
