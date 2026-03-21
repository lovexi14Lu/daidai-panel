package model_test

import (
	"testing"

	"daidai-panel/database"
	"daidai-panel/model"
	"daidai-panel/testutil"
)

func TestSetConfigNormalizesRegisteredValues(t *testing.T) {
	testutil.SetupTestEnv(t)

	if err := model.SetConfig("auto_install_deps", "0"); err != nil {
		t.Fatalf("set auto_install_deps: %v", err)
	}
	if got := model.GetRegisteredConfigBool("auto_install_deps"); got {
		t.Fatalf("expected auto_install_deps to be false after normalization")
	}

	var autoInstall model.SystemConfig
	if err := database.DB.Where("`key` = ?", "auto_install_deps").First(&autoInstall).Error; err != nil {
		t.Fatalf("query auto_install_deps: %v", err)
	}
	if autoInstall.Value != "false" {
		t.Fatalf("expected canonical bool value false, got %q", autoInstall.Value)
	}

	if err := model.SetConfig("captcha_fail_mode", " strict "); err != nil {
		t.Fatalf("set captcha_fail_mode: %v", err)
	}
	if got := model.GetRegisteredConfig("captcha_fail_mode"); got != "strict" {
		t.Fatalf("expected captcha_fail_mode strict, got %q", got)
	}

	if err := model.SetConfig("trusted_proxy_cidrs", "127.0.0.1, 203.0.113.10"); err != nil {
		t.Fatalf("set trusted_proxy_cidrs: %v", err)
	}
	if got := model.GetRegisteredConfig("trusted_proxy_cidrs"); got != "127.0.0.1/32\n203.0.113.10/32" {
		t.Fatalf("expected canonical trusted_proxy_cidrs, got %q", got)
	}

	if err := model.SetConfig("default_cron_rule", "invalid cron"); err == nil {
		t.Fatal("expected invalid default_cron_rule to be rejected")
	}
	if err := model.SetConfig("trusted_proxy_cidrs", "not-an-ip"); err == nil {
		t.Fatal("expected invalid trusted_proxy_cidrs to be rejected")
	}
}

func TestRegisteredConfigUsesRegistryDefaults(t *testing.T) {
	testutil.SetupTestEnv(t)

	database.DB.Where("`key` = ?", "panel_title").Delete(&model.SystemConfig{})

	if got := model.GetRegisteredConfig("panel_title"); got != "呆呆面板" {
		t.Fatalf("expected registry default panel_title, got %q", got)
	}
	if got := model.GetRegisteredConfigInt("command_timeout"); got != 300 {
		t.Fatalf("expected registry default command_timeout 300, got %d", got)
	}
	if got := model.GetRegisteredConfigBool("notify_on_login"); got {
		t.Fatalf("expected registry default notify_on_login to be false")
	}
}
