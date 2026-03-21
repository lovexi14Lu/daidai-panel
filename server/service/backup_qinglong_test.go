package service

import (
	"encoding/json"
	"testing"
)

func TestMapQingLongConfigToSystemConfig(t *testing.T) {
	key, value, ok := mapQingLongConfigToSystemConfig("CommandTimeoutTime", "1h")
	if !ok {
		t.Fatal("expected command timeout to be mapped")
	}
	if key != "command_timeout" {
		t.Fatalf("expected command_timeout key, got %q", key)
	}
	if value != "3600" {
		t.Fatalf("expected 3600 seconds, got %q", value)
	}
}

func TestMapQingLongDependencyType(t *testing.T) {
	if got := mapQingLongDependencyType(0); got != "nodejs" {
		t.Fatalf("expected nodejs, got %q", got)
	}
	if got := mapQingLongDependencyType(1); got != "python" {
		t.Fatalf("expected python, got %q", got)
	}
	if got := mapQingLongDependencyType(99); got != "" {
		t.Fatalf("expected empty mapping for unknown type, got %q", got)
	}
}

func TestBuildQingLongNotificationChannels(t *testing.T) {
	channels := buildQingLongNotificationChannels(map[string]string{
		"PUSH_KEY":           "SCT123456",
		"DD_BOT_TOKEN":       "ding-token",
		"DD_BOT_SECRET":      "ding-secret",
		"QYWX_KEY":           "qywx-key",
		"BARK_PUSH":          "https://api.day.app/device-key",
		"DEER_KEY":           "pushdeer-key",
		"DEER_URL":           "https://api2.pushdeer.com",
		"PUSHME_KEY":         "pushme-key",
		"PUSHME_URL":         "https://push.i-i.me/",
		"QMSG_KEY":           "qmsg-key",
		"QMSG_TYPE":          "group",
		"WEBHOOK_URL":        "https://example.com/webhook",
		"WEBHOOK_HEADERS":    "Authorization: Bearer demo\nX-Test: 1",
		"WEBHOOK_METHOD":     "POST",
		"WEBHOOK_BODY":       "{\"msg\":\"{{title}}\"}",
		"FSKEY":              "feishu-key",
		"FSSECRET":           "feishu-secret",
		"NTFY_TOPIC":         "demo-topic",
		"NTFY_URL":           "https://ntfy.sh",
		"NTFY_PRIORITY":      "4",
		"NTFY_TOKEN":         "secret-token",
		"PUSH_PLUS_TOKEN":    "pushplus-token",
		"PUSH_PLUS_USER":     "group-1",
		"PUSH_PLUS_TEMPLATE": "markdown",
		"WXPUSHER_APP_TOKEN": "wxpusher-token",
		"WXPUSHER_TOPIC_IDS": "101;102",
		"WXPUSHER_UIDS":      "UID_demo_1;UID_demo_2",
	})

	byType := make(map[string]map[string]string, len(channels))
	for _, channel := range channels {
		var cfg map[string]string
		if err := json.Unmarshal([]byte(channel.Config), &cfg); err != nil {
			t.Fatalf("unmarshal %s config: %v", channel.Type, err)
		}
		byType[channel.Type] = cfg
	}

	if got := byType["serverchan"]["key"]; got != "SCT123456" {
		t.Fatalf("expected serverchan key, got %q", got)
	}
	if got := byType["dingtalk"]["webhook"]; got != "https://oapi.dingtalk.com/robot/send?access_token=ding-token" {
		t.Fatalf("unexpected dingtalk webhook: %q", got)
	}
	if got := byType["wecom"]["webhook"]; got != "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=qywx-key" {
		t.Fatalf("unexpected wecom webhook: %q", got)
	}
	if got := byType["bark"]["key"]; got != "device-key" {
		t.Fatalf("unexpected bark key: %q", got)
	}
	if got := byType["pushdeer"]["server"]; got != "https://api2.pushdeer.com" {
		t.Fatalf("unexpected pushdeer server: %q", got)
	}
	if got := byType["pushme"]["server"]; got != "https://push.i-i.me/" {
		t.Fatalf("unexpected pushme server: %q", got)
	}
	if got := byType["qmsg"]["mode"]; got != "group" {
		t.Fatalf("unexpected qmsg mode: %q", got)
	}
	if got := byType["custom"]["headers"]; got == "" {
		t.Fatal("expected custom webhook headers to be normalized into JSON")
	}
	if got := byType["feishu"]["secret"]; got != "feishu-secret" {
		t.Fatalf("unexpected feishu secret: %q", got)
	}
	if got := byType["ntfy"]["topic"]; got != "demo-topic" {
		t.Fatalf("unexpected ntfy topic: %q", got)
	}
	if got := byType["pushplus"]["topic"]; got != "group-1" {
		t.Fatalf("unexpected pushplus topic: %q", got)
	}
	if got := byType["wxpusher"]["app_token"]; got != "wxpusher-token" {
		t.Fatalf("unexpected wxpusher app token: %q", got)
	}
	if got := byType["wxpusher"]["topic_ids"]; got != "101;102" {
		t.Fatalf("unexpected wxpusher topic ids: %q", got)
	}
	if got := byType["wxpusher"]["uids"]; got != "UID_demo_1;UID_demo_2" {
		t.Fatalf("unexpected wxpusher uids: %q", got)
	}
	if got := byType["wxpusher"]["content_type"]; got != "2" {
		t.Fatalf("unexpected wxpusher content type: %q", got)
	}
}
