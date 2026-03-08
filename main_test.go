package main

import (
	"backupgo/config"
	"testing"
)

func before() {
	config.InitConfig()
}

func Test_backup(t *testing.T) {
	t.Skip("integration test requires local config and external services")

	before()

	c := defaultHolder("test", config.BackupConfig{
		BackPath: "~/Downloads/MapleMonoNormalNL-TTF",
	})

	c.initLogger()
	if err := c.backupWithLogger(); err != nil {
		t.Fatal(err)
	}
	c.sendMessages()
}

func Test_cleanOld(t *testing.T) {
	t.Skip("integration test requires local config and external services")

	before()

	th := defaultHolder("test", config.BackupConfig{
		BackPath: "E:/audio/asmr",
	})
	if err := th.cleanHistoryWithLogger(); err != nil {
		t.Fatal(err)
	}
}
