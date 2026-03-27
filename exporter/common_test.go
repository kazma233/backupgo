package exporter

import (
	"backupgo/config"
	"reflect"
	"testing"
)

func TestNewSource(t *testing.T) {
	tests := []struct {
		name     string
		conf     config.BackupConfig
		wantType string
	}{
		{
			name: "path",
			conf: config.BackupConfig{
				ID:         "path",
				BackupPath: "./export",
			},
			wantType: "exporter.pathSource",
		},
		{
			name: "postgres",
			conf: config.BackupConfig{
				ID:   "pg",
				Type: config.BackupTypePostgres,
				Postgres: &config.PostgresBackupConfig{
					Databases: []string{"app"},
				},
			},
			wantType: "exporter.postgresBackupSource",
		},
		{
			name: "mongodb",
			conf: config.BackupConfig{
				ID:   "mongo",
				Type: config.BackupTypeMongoDB,
				MongoDB: &config.MongoBackupConfig{
					Databases: []string{"app"},
				},
			},
			wantType: "exporter.mongoBackupSource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := New(tt.conf)
			if err != nil {
				t.Fatalf("New returned error: %v", err)
			}
			if got := reflect.TypeOf(source).String(); got != tt.wantType {
				t.Fatalf("unexpected source type: %s", got)
			}
		})
	}
}
