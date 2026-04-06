package appboot

import (
	"fmt"
	"os"

	"daidai-panel/config"
	"daidai-panel/database"
	"daidai-panel/middleware"
	"daidai-panel/model"
)

func ResolveConfigPath() string {
	candidates := []string{
		os.Getenv("DAIDAI_CONFIG"),
		"/app/config.yaml",
		"config.yaml",
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return "config.yaml"
}

func LoadAndInit(configPath string) (*config.Config, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	if err := InitWithConfig(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func InitWithConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("配置为空")
	}

	database.Init(&cfg.Database)
	database.AutoMigrate(allModels()...)
	database.EnsureColumns()

	model.InitDefaultConfigs()
	if err := middleware.ConfigureTrustedProxyCIDRs(model.GetRegisteredConfig("trusted_proxy_cidrs")); err != nil {
		return fmt.Errorf("failed to configure trusted proxies: %w", err)
	}

	return nil
}

func allModels() []interface{} {
	return []interface{}{
		&model.User{},
		&model.TokenBlocklist{},
		&model.Task{},
		&model.TaskLog{},
		&model.SystemConfig{},
		&model.EnvVar{},
		&model.ScriptVersion{},
		&model.Subscription{},
		&model.SubLog{},
		&model.NotifyChannel{},
		&model.SSHKey{},
		&model.LoginLog{},
		&model.LoginAttempt{},
		&model.UserSession{},
		&model.IPWhitelist{},
		&model.SecurityAudit{},
		&model.TwoFactorAuth{},
		&model.OpenApp{},
		&model.ApiCallLog{},
		&model.Platform{},
		&model.PlatformToken{},
		&model.PlatformTokenLog{},
		&model.Dependency{},
		&model.TaskView{},
	}
}
