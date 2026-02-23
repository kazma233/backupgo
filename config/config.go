package config

import (
	"io"
	"os"

	"github.com/goccy/go-yaml"
)

type (
	// GlobalConfig base config
	GlobalConfig struct {
		OSS        OssConfig               `yaml:"oss"`
		Mail       *MailConfig             `yaml:"mail"`
		TG         *TGConfig               `yaml:"tg"`
		TgChatId   string                  `yaml:"tg_chat_id"`
		NoticeMail []string                `yaml:"notice_mail"`
		BackupConf map[string]BackupConfig `yaml:"backup"`
	}

	BackupConfig struct {
		BeforeCmd  string `yaml:"before_command"`
		BackPath   string `yaml:"back_path"`
		AfterCmd   string `yaml:"after_command"`
		BackupTask string `yaml:"backup_task"`
	}

	OssConfig struct {
		BucketName      string `yaml:"bucket_name"`
		AccessKey       string `yaml:"access_key"`
		AccessKeySecret string `yaml:"access_key_secret"`
		Endpoint        string `yaml:"endpoint"`
		FastEndpoint    string `yaml:"fast_endpoint"`
	}

	TGConfig struct {
		Key string `yaml:"key"`
	}

	MailConfig struct {
		Smtp     string `yaml:"smtp"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	}
)

var (
	Config GlobalConfig
)

func InitConfig() {
	f, err := os.OpenFile("config.yml", os.O_RDONLY, 0755)
	if err != nil {
		panic(err)
	}

	configBlob, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	var config = GlobalConfig{}
	err = yaml.Unmarshal(configBlob, &config)
	if err != nil {
		panic(err)
	}

	if len(config.BackupConf) <= 0 {
		panic("config can not be empty")
	}

	for _, v := range config.BackupConf {
		if v.BackPath == "" {
			panic("id or back_path can not be empty")
		}
	}

	Config = config
}
