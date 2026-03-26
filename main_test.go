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

	c := newTaskHolder(config.BackupConfig{
		ID:         "test",
		BackupPath: "~/Downloads/MapleMonoNormalNL-TTF",
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

	th := newTaskHolder(config.BackupConfig{
		ID:         "test",
		BackupPath: "E:/audio/asmr",
	})
	if err := th.cleanHistoryWithLogger(); err != nil {
		t.Fatal(err)
	}
}
