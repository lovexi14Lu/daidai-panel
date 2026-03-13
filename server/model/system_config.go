package model

import (
	"errors"
	"time"

	"daidai-panel/database"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SystemConfig struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Key         string    `gorm:"size:64;uniqueIndex;not null" json:"key"`
	Value       string    `gorm:"type:text;default:''" json:"value"`
	Description string    `gorm:"size:256;default:''" json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}

func silentDB() *gorm.DB {
	return database.DB.Session(&gorm.Session{Logger: database.DB.Logger.LogMode(logger.Silent)})
}

func GetConfig(key string, defaultValue string) string {
	var cfg SystemConfig
	if err := silentDB().Where("`key` = ?", key).First(&cfg).Error; err != nil {
		return defaultValue
	}
	if cfg.Value == "" {
		return defaultValue
	}
	return cfg.Value
}

func GetConfigInt(key string, defaultValue int) int {
	val := GetConfig(key, "")
	if val == "" {
		return defaultValue
	}
	var result int
	for _, c := range val {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			return defaultValue
		}
	}
	return result
}

func SetConfig(key, value string) error {
	var cfg SystemConfig
	if err := silentDB().Where("`key` = ?", key).First(&cfg).Error; err != nil {
		cfg = SystemConfig{Key: key, Value: value}
		return database.DB.Create(&cfg).Error
	}
	return database.DB.Model(&cfg).Update("value", value).Error
}

func InitDefaultConfigs() {
	defaults := map[string]struct {
		Value       string
		Description string
	}{
		"max_concurrent_tasks":    {"5", "定时任务最大并发数"},
		"command_timeout":         {"300", "全局默认超时（秒）"},
		"log_retention_days":      {"7", "日志保留天数"},
		"cpu_warn":                {"80", "CPU 告警阈值（%）"},
		"memory_warn":             {"80", "内存告警阈值（%）"},
		"disk_warn":               {"90", "磁盘告警阈值（%）"},
		"auto_add_cron":           {"true", "自动添加定时任务"},
		"auto_del_cron":           {"true", "自动删除失效任务"},
		"notify_on_resource_warn": {"false", "资源超限发送通知"},
		"notify_on_login":         {"false", "登录成功发送通知"},
	}

	db := silentDB()
	for key, cfg := range defaults {
		var existing SystemConfig
		if err := db.Where("`key` = ?", key).First(&existing).Error; err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			database.DB.Create(&SystemConfig{
				Key:         key,
				Value:       cfg.Value,
				Description: cfg.Description,
			})
		}
	}
}
