package main

import (
	"backupgo/config"
	"testing"
)

func before() {
	config.InitConfig()
}

func Test_backup(t *testing.T) {
	before()

	c := defaultHolder("test", config.BackupConfig{
		BackPath: "~/Downloads/MapleMonoNormalNL-TTF",
	})

	c.initLogger()
	c.backupWithLogger()
	c.sendMessages()
}

func Test_cleanOld(t *testing.T) {
	before()

	th := defaultHolder("test", config.BackupConfig{
		BackPath: "E:/audio/asmr",
	})
	th.cleanHistory()
}
