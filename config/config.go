package config

import (
	"errors"
	"io"
	"os"

	"github.com/goccy/go-yaml"
)

type (
	// GlobalConfig base config
	GlobalConfig struct {
		OSS        OssConfig               `yaml:"oss"`
		Notice     *NoticeConfig           `yaml:"notice"`
		BackupConf map[string]BackupConfig `yaml:"backup"`
	}

	NoticeConfig struct {
		Mail     *MailConfig     `yaml:"mail"`
		Telegram *TelegramConfig `yaml:"telegram"`
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

	TelegramConfig struct {
		BotToken string `yaml:"bot_token"`
		ChatID   string `yaml:"chat_id"`
	}

	MailConfig struct {
		Smtp     string   `yaml:"smtp"`
		Port     int      `yaml:"port"`
		User     string   `yaml:"user"`
		Password string   `yaml:"password"`
		To       []string `yaml:"to"`
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
	defer f.Close()

	configBlob, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	config, err := ParseConfig(configBlob)
	if err != nil {
		panic(err)
	}

	Config = config
}

func ParseConfig(configBlob []byte) (GlobalConfig, error) {
	var config GlobalConfig
	if err := yaml.Unmarshal(configBlob, &config); err != nil {
		return GlobalConfig{}, err
	}

	if len(config.BackupConf) <= 0 {
		return GlobalConfig{}, errors.New("config can not be empty")
	}

	for _, v := range config.BackupConf {
		if v.BackPath == "" {
			return GlobalConfig{}, errors.New("id or back_path can not be empty")
		}
	}

	return config, nil
}
