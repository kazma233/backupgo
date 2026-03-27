package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

const (
	BackupTypePath         = "path"
	BackupTypePostgres     = "postgres"
	BackupTypeMongoDB      = "mongodb"
	BackupTypeDockerVolume = "docker_volume"

	ExecModeLocal  = "local"
	ExecModeDocker = "docker"
)

type (
	// GlobalConfig base config
	GlobalConfig struct {
		OSS        OssConfig      `yaml:"oss"`
		Notice     *NoticeConfig  `yaml:"notice"`
		BackupConf []BackupConfig `yaml:"backup"`
	}

	NoticeConfig struct {
		Mail     *MailConfig     `yaml:"mail"`
		Telegram *TelegramConfig `yaml:"telegram"`
	}

	BackupConfig struct {
		ID           string                    `yaml:"id"`
		Type         string                    `yaml:"type"`
		BeforeCmd    string                    `yaml:"before_command"`
		BackupPath   string                    `yaml:"backup_path"`
		AfterCmd     string                    `yaml:"after_command"`
		BackupTask   string                    `yaml:"backup_task"`
		Postgres     *PostgresBackupConfig     `yaml:"postgres"`
		MongoDB      *MongoBackupConfig        `yaml:"mongodb"`
		DockerVolume *DockerVolumeBackupConfig `yaml:"docker_volume"`
	}

	PostgresBackupConfig struct {
		Mode      string   `yaml:"mode"`
		Container string   `yaml:"container"`
		Host      string   `yaml:"host"`
		Port      int      `yaml:"port"`
		User      string   `yaml:"user"`
		Password  string   `yaml:"password"`
		Databases []string `yaml:"databases"`
		ExtraArgs []string `yaml:"extra_args"`
	}

	MongoBackupConfig struct {
		Mode         string   `yaml:"mode"`
		Container    string   `yaml:"container"`
		URI          string   `yaml:"uri"`
		Host         string   `yaml:"host"`
		Port         int      `yaml:"port"`
		Username     string   `yaml:"username"`
		Password     string   `yaml:"password"`
		AuthDatabase string   `yaml:"auth_database"`
		Databases    []string `yaml:"databases"`
		Gzip         bool     `yaml:"gzip"`
		ExtraArgs    []string `yaml:"extra_args"`
	}

	DockerVolumeBackupConfig struct {
		Volume string `yaml:"volume"`
		Image  string `yaml:"image"`
	}

	OssConfig struct {
		BucketName      string `yaml:"bucket_name"`
		AccessKey       string `yaml:"access_key"`
		AccessKeySecret string `yaml:"access_key_secret"`
		Region          string `yaml:"region"`
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

func (c BackupConfig) GetID() string {
	return strings.TrimSpace(c.ID)
}

func (c BackupConfig) GetType() string {
	if normalized := strings.ToLower(strings.TrimSpace(c.Type)); normalized != "" {
		return normalized
	}
	if c.Postgres != nil {
		return BackupTypePostgres
	}
	if c.MongoDB != nil {
		return BackupTypeMongoDB
	}
	if c.DockerVolume != nil {
		return BackupTypeDockerVolume
	}
	return BackupTypePath
}

func (c BackupConfig) Validate() error {
	taskID := c.GetID()
	if strings.TrimSpace(taskID) == "" {
		return errors.New("id can not be empty")
	}

	sourceCount := 0
	if strings.TrimSpace(c.BackupPath) != "" {
		sourceCount++
	}
	if c.Postgres != nil {
		sourceCount++
	}
	if c.MongoDB != nil {
		sourceCount++
	}
	if c.DockerVolume != nil {
		sourceCount++
	}

	if sourceCount == 0 {
		return fmt.Errorf("backup %s must configure one source", taskID)
	}
	if sourceCount > 1 {
		return fmt.Errorf("backup %s only supports one source at a time", taskID)
	}

	switch c.GetType() {
	case BackupTypePath:
		if strings.TrimSpace(c.BackupPath) == "" {
			return fmt.Errorf("backup %s backup_path can not be empty", taskID)
		}
		if c.Postgres != nil || c.MongoDB != nil || c.DockerVolume != nil {
			return fmt.Errorf("backup %s path source can not be combined with another source", taskID)
		}
	case BackupTypePostgres:
		if c.Postgres == nil {
			return fmt.Errorf("backup %s postgres config can not be empty", taskID)
		}
		if err := c.Postgres.Validate(taskID); err != nil {
			return err
		}
	case BackupTypeMongoDB:
		if c.MongoDB == nil {
			return fmt.Errorf("backup %s mongodb config can not be empty", taskID)
		}
		if err := c.MongoDB.Validate(taskID); err != nil {
			return err
		}
	case BackupTypeDockerVolume:
		if c.DockerVolume == nil {
			return fmt.Errorf("backup %s docker_volume config can not be empty", taskID)
		}
		if err := c.DockerVolume.Validate(taskID); err != nil {
			return err
		}
	default:
		return fmt.Errorf("backup %s has unsupported type %q", taskID, c.Type)
	}

	return nil
}

func (g GlobalConfig) FindBackupByID(id string) (BackupConfig, bool) {
	targetID := strings.TrimSpace(id)
	for _, conf := range g.BackupConf {
		if conf.GetID() == targetID {
			return conf, true
		}
	}
	return BackupConfig{}, false
}

func (g GlobalConfig) BackupIDs() []string {
	ids := make([]string, 0, len(g.BackupConf))
	for _, conf := range g.BackupConf {
		ids = append(ids, conf.GetID())
	}
	return ids
}

func (c PostgresBackupConfig) Validate(taskID string) error {
	if len(c.Databases) == 0 {
		return fmt.Errorf("backup %s postgres.databases can not be empty", taskID)
	}
	mode := c.GetMode()
	if mode != ExecModeLocal && mode != ExecModeDocker {
		return fmt.Errorf("backup %s postgres.mode must be one of %q or %q", taskID, ExecModeLocal, ExecModeDocker)
	}
	if mode == ExecModeDocker && strings.TrimSpace(c.Container) == "" {
		return fmt.Errorf("backup %s postgres.container can not be empty when mode is docker", taskID)
	}

	return nil
}

func (c PostgresBackupConfig) GetMode() string {
	mode := strings.ToLower(strings.TrimSpace(c.Mode))
	if mode == "" {
		return ExecModeLocal
	}
	return mode
}

func (c MongoBackupConfig) Validate(taskID string) error {
	if len(c.Databases) == 0 {
		return fmt.Errorf("backup %s mongodb.databases can not be empty", taskID)
	}
	mode := c.GetMode()
	if mode != ExecModeLocal && mode != ExecModeDocker {
		return fmt.Errorf("backup %s mongodb.mode must be one of %q or %q", taskID, ExecModeLocal, ExecModeDocker)
	}
	if mode == ExecModeDocker && strings.TrimSpace(c.Container) == "" {
		return fmt.Errorf("backup %s mongodb.container can not be empty when mode is docker", taskID)
	}
	if strings.TrimSpace(c.URI) == "" && strings.TrimSpace(c.Username) != "" && c.Password == "" {
		return fmt.Errorf("backup %s mongodb.password can not be empty when username is set", taskID)
	}

	return nil
}

func (c MongoBackupConfig) GetMode() string {
	mode := strings.ToLower(strings.TrimSpace(c.Mode))
	if mode == "" {
		return ExecModeLocal
	}
	return mode
}

func (c DockerVolumeBackupConfig) Validate(taskID string) error {
	if strings.TrimSpace(c.Volume) == "" {
		return fmt.Errorf("backup %s docker_volume.volume can not be empty", taskID)
	}

	return nil
}

func (c DockerVolumeBackupConfig) GetImage() string {
	image := strings.TrimSpace(c.Image)
	if image == "" {
		return "busybox:latest"
	}

	return image
}

func InitConfig() {
	configBlob, err := os.ReadFile("config.yml")
	if err != nil {
		configBlob, err = os.ReadFile("config.yaml")
		if err != nil {
			panic(err)
		}
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

	seenIDs := make(map[string]struct{}, len(config.BackupConf))
	for _, v := range config.BackupConf {
		if err := v.Validate(); err != nil {
			return GlobalConfig{}, err
		}

		id := v.GetID()
		if _, exists := seenIDs[id]; exists {
			return GlobalConfig{}, fmt.Errorf("duplicate backup id: %s", id)
		}
		seenIDs[id] = struct{}{}
	}

	return config, nil
}
