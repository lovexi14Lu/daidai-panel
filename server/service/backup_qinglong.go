package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"daidai-panel/config"
	"daidai-panel/model"
	panelcron "daidai-panel/pkg/cron"

	_ "github.com/glebarez/sqlite"
)

func buildQingLongManifest(extractedDir string) (BackupManifest, error) {
	dataDir, err := resolveQingLongDataDir(extractedDir)
	if err != nil {
		return BackupManifest{}, err
	}

	manifest := BackupManifest{
		Format:    "daidai-panel-backup",
		Version:   "0.4.0",
		Source:    "qinglong",
		CreatedAt: time.Now(),
	}

	configPath := filepath.Join(dataDir, "config", "config.sh")
	if _, err := os.Stat(configPath); err == nil {
		configs, envs, channels, err := parseQingLongConfig(configPath)
		if err != nil {
			return BackupManifest{}, err
		}
		if len(configs) > 0 {
			manifest.Selection.Configs = true
			manifest.Data.Configs.SystemConfigs = append(manifest.Data.Configs.SystemConfigs, configs...)
		}
		if len(channels) > 0 {
			manifest.Selection.Configs = true
			manifest.Data.Configs.NotifyChannels = append(manifest.Data.Configs.NotifyChannels, channels...)
		}
		if len(envs) > 0 {
			manifest.Selection.EnvVars = true
			manifest.Data.EnvVars = append(manifest.Data.EnvVars, envs...)
		}
	}

	dbPath := filepath.Join(dataDir, "db", "database.sqlite")
	if _, err := os.Stat(dbPath); err == nil {
		if err := enrichManifestFromQingLongDB(dbPath, &manifest); err != nil {
			return BackupManifest{}, err
		}
	}

	if hasAnyQingLongScriptFiles(dataDir) {
		manifest.Selection.Scripts = true
	}
	if hasAnyQingLongLogFiles(dataDir) {
		manifest.Selection.Logs = true
	}
	if !manifest.Selection.Any() {
		return BackupManifest{}, fmt.Errorf("未识别到可导入的青龙备份内容")
	}

	return manifest, nil
}

func resolveQingLongDataDir(extractedDir string) (string, error) {
	candidates := []string{
		filepath.Join(extractedDir, "data"),
		extractedDir,
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(filepath.Join(candidate, "config")); err == nil {
			if _, err := os.Stat(filepath.Join(candidate, "db")); err == nil {
				return candidate, nil
			}
		}
	}
	return "", fmt.Errorf("未检测到青龙备份目录结构")
}

func parseQingLongConfig(configPath string) ([]model.SystemConfig, []model.EnvVar, []BackupNotifyChannel, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("读取青龙配置失败: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	configs := make([]model.SystemConfig, 0)
	envs := make([]model.EnvVar, 0)
	exportedVars := make(map[string]string)
	seenConfigKeys := map[string]struct{}{}

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		isExport := false
		if strings.HasPrefix(line, "export ") {
			isExport = true
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := normalizeQingLongShellValue(parts[1])
		if key == "" {
			continue
		}

		if isExport {
			if strings.TrimSpace(value) == "" {
				continue
			}
			exportedVars[key] = value
			envs = append(envs, model.EnvVar{
				Name:      key,
				Value:     value,
				Enabled:   true,
				Position:  float64(len(envs) + 1),
				SortOrder: len(envs),
			})
			continue
		}

		mappedKey, mappedValue, ok := mapQingLongConfigToSystemConfig(key, value)
		if !ok {
			continue
		}
		if _, exists := seenConfigKeys[mappedKey]; exists {
			continue
		}
		seenConfigKeys[mappedKey] = struct{}{}
		configs = append(configs, model.SystemConfig{Key: mappedKey, Value: mappedValue})
	}

	channels := buildQingLongNotificationChannels(exportedVars)

	return configs, envs, channels, nil
}

func normalizeQingLongShellValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	return strings.TrimSpace(value)
}

func mapQingLongConfigToSystemConfig(key, value string) (string, string, bool) {
	switch key {
	case "AutoAddCron":
		return buildNormalizedSystemConfig("auto_add_cron", value)
	case "AutoDelCron":
		return buildNormalizedSystemConfig("auto_del_cron", value)
	case "DefaultCronRule":
		return buildNormalizedSystemConfig("default_cron_rule", value)
	case "RepoFileExtensions":
		return buildNormalizedSystemConfig("repo_file_extensions", value)
	case "ProxyUrl":
		return buildNormalizedSystemConfig("proxy_url", value)
	case "CpuWarn":
		return buildNormalizedSystemConfig("cpu_warn", value)
	case "MemoryWarn":
		return buildNormalizedSystemConfig("memory_warn", value)
	case "DiskWarn":
		return buildNormalizedSystemConfig("disk_warn", value)
	case "RandomDelay":
		return buildNormalizedSystemConfig("random_delay", value)
	case "RandomDelayFileExtensions":
		return buildNormalizedSystemConfig("random_delay_extensions", value)
	case "CommandTimeoutTime":
		seconds, ok := parseQingLongDurationSeconds(value)
		if !ok {
			return "", "", false
		}
		return buildNormalizedSystemConfig("command_timeout", strconv.Itoa(seconds))
	default:
		return "", "", false
	}
}

func buildNormalizedSystemConfig(key, value string) (string, string, bool) {
	normalized, err := model.NormalizeSystemConfigValue(key, value)
	if err != nil {
		return "", "", false
	}
	return key, normalized, true
}

func parseQingLongDurationSeconds(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if strings.HasSuffix(value, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(value, "d"))
		if err != nil || days <= 0 {
			return 0, false
		}
		return days * 24 * 3600, true
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, false
	}
	return int(duration.Seconds()), true
}

func buildQingLongNotificationChannels(env map[string]string) []BackupNotifyChannel {
	channels := make([]BackupNotifyChannel, 0)

	appendChannel := func(name, channelType string, cfg map[string]string) {
		if len(cfg) == 0 {
			return
		}
		configJSON, err := json.Marshal(cfg)
		if err != nil {
			return
		}
		channels = append(channels, BackupNotifyChannel{
			Name:    name,
			Type:    channelType,
			Config:  string(configJSON),
			Enabled: true,
		})
	}

	if key := strings.TrimSpace(env["PUSH_KEY"]); key != "" {
		appendChannel("青龙导入 - Server酱", "serverchan", map[string]string{
			"key": key,
		})
	}

	if barkCfg := buildQingLongBarkConfig(env); len(barkCfg) > 0 {
		appendChannel("青龙导入 - Bark", "bark", barkCfg)
	}

	if token := strings.TrimSpace(env["TG_BOT_TOKEN"]); token != "" {
		if chatID := strings.TrimSpace(env["TG_USER_ID"]); chatID != "" {
			cfg := map[string]string{
				"token":   token,
				"chat_id": chatID,
			}
			if apiHost := strings.TrimSpace(env["TG_API_HOST"]); apiHost != "" {
				if !strings.HasPrefix(apiHost, "http://") && !strings.HasPrefix(apiHost, "https://") {
					apiHost = "https://" + apiHost
				}
				cfg["api_host"] = apiHost
			}
			appendChannel("青龙导入 - Telegram", "telegram", cfg)
		}
	}

	if token := strings.TrimSpace(env["DD_BOT_TOKEN"]); token != "" {
		cfg := map[string]string{
			"webhook": fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s", token),
		}
		if secret := strings.TrimSpace(env["DD_BOT_SECRET"]); secret != "" {
			cfg["secret"] = secret
		}
		appendChannel("青龙导入 - 钉钉", "dingtalk", cfg)
	}

	if key := strings.TrimSpace(env["QYWX_KEY"]); key != "" {
		appendChannel("青龙导入 - 企业微信机器人", "wecom", map[string]string{
			"webhook": fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=%s", key),
		})
	}

	if token := strings.TrimSpace(env["PUSH_PLUS_TOKEN"]); token != "" {
		cfg := map[string]string{
			"token": token,
		}
		if topic := strings.TrimSpace(env["PUSH_PLUS_USER"]); topic != "" {
			cfg["topic"] = topic
		}
		if template := strings.TrimSpace(env["PUSH_PLUS_TEMPLATE"]); template != "" {
			cfg["template"] = template
		}
		appendChannel("青龙导入 - PushPlus", "pushplus", cfg)
	}

	if gotifyURL := strings.TrimSpace(env["GOTIFY_URL"]); gotifyURL != "" {
		if token := strings.TrimSpace(env["GOTIFY_TOKEN"]); token != "" {
			cfg := map[string]string{
				"server": gotifyURL,
				"token":  token,
			}
			if priority := strings.TrimSpace(env["GOTIFY_PRIORITY"]); priority != "" {
				cfg["priority"] = priority
			}
			appendChannel("青龙导入 - Gotify", "gotify", cfg)
		}
	}

	if key := strings.TrimSpace(env["DEER_KEY"]); key != "" {
		cfg := map[string]string{
			"key": key,
		}
		if server := strings.TrimSpace(env["DEER_URL"]); server != "" {
			cfg["server"] = server
		}
		appendChannel("青龙导入 - PushDeer", "pushdeer", cfg)
	}

	if key := strings.TrimSpace(env["PUSHME_KEY"]); key != "" {
		cfg := map[string]string{
			"key": key,
		}
		if server := strings.TrimSpace(env["PUSHME_URL"]); server != "" {
			cfg["server"] = server
		}
		appendChannel("青龙导入 - PushMe", "pushme", cfg)
	}

	if key := strings.TrimSpace(env["IGOT_PUSH_KEY"]); key != "" {
		appendChannel("青龙导入 - iGot", "igot", map[string]string{
			"key": key,
		})
	}

	if key := strings.TrimSpace(env["QMSG_KEY"]); key != "" {
		cfg := map[string]string{
			"key": key,
		}
		if mode := strings.ToLower(strings.TrimSpace(env["QMSG_TYPE"])); mode != "" {
			cfg["mode"] = mode
		}
		appendChannel("青龙导入 - Qmsg", "qmsg", cfg)
	}

	if key := strings.TrimSpace(env["FSKEY"]); key != "" {
		cfg := map[string]string{
			"webhook": fmt.Sprintf("https://open.feishu.cn/open-apis/bot/v2/hook/%s", key),
		}
		if secret := strings.TrimSpace(env["FSSECRET"]); secret != "" {
			cfg["secret"] = secret
		}
		appendChannel("青龙导入 - 飞书", "feishu", cfg)
	}

	if topic := strings.TrimSpace(env["NTFY_TOPIC"]); topic != "" {
		cfg := map[string]string{
			"topic": topic,
		}
		if server := strings.TrimSpace(env["NTFY_URL"]); server != "" {
			cfg["server"] = server
		}
		if priority := strings.TrimSpace(env["NTFY_PRIORITY"]); priority != "" {
			cfg["priority"] = priority
		}
		if token := strings.TrimSpace(env["NTFY_TOKEN"]); token != "" {
			cfg["token"] = token
		}
		appendChannel("青龙导入 - ntfy", "ntfy", cfg)
	}

	if appToken := strings.TrimSpace(env["WXPUSHER_APP_TOKEN"]); appToken != "" {
		cfg := map[string]string{
			"app_token":    appToken,
			"content_type": "2",
		}
		copyIfNotEmpty(cfg, "topic_ids", env["WXPUSHER_TOPIC_IDS"])
		copyIfNotEmpty(cfg, "uids", env["WXPUSHER_UIDS"])
		if cfg["topic_ids"] != "" || cfg["uids"] != "" {
			appendChannel("青龙导入 - WxPusher", "wxpusher", cfg)
		}
	}

	if customCfg := buildQingLongCustomWebhookConfig(env); len(customCfg) > 0 {
		appendChannel("青龙导入 - 自定义通知", "custom", customCfg)
	}

	if emailCfg := buildQingLongEmailConfig(env); len(emailCfg) > 0 {
		appendChannel("青龙导入 - SMTP 邮件", "email", emailCfg)
	}

	return channels
}

func buildQingLongBarkConfig(env map[string]string) map[string]string {
	pushValue := strings.TrimSpace(env["BARK_PUSH"])
	if pushValue == "" {
		return nil
	}

	cfg := map[string]string{}
	if strings.HasPrefix(pushValue, "http://") || strings.HasPrefix(pushValue, "https://") {
		if parsed, err := url.Parse(pushValue); err == nil {
			key := strings.Trim(strings.TrimSpace(parsed.Path), "/")
			if key != "" {
				cfg["key"] = filepath.Base(key)
				serverPath := strings.TrimSuffix(parsed.Path, "/"+cfg["key"])
				cfg["server"] = strings.TrimRight(parsed.Scheme+"://"+parsed.Host+serverPath, "/")
			}
		}
	}
	if cfg["key"] == "" {
		cfg["key"] = pushValue
	}
	if cfg["server"] == "" {
		delete(cfg, "server")
	}

	copyIfNotEmpty(cfg, "icon", env["BARK_ICON"])
	copyIfNotEmpty(cfg, "sound", env["BARK_SOUND"])
	copyIfNotEmpty(cfg, "group", env["BARK_GROUP"])
	copyIfNotEmpty(cfg, "level", env["BARK_LEVEL"])
	copyIfNotEmpty(cfg, "url", env["BARK_URL"])

	return cfg
}

func buildQingLongCustomWebhookConfig(env map[string]string) map[string]string {
	webhookURL := strings.TrimSpace(env["WEBHOOK_URL"])
	if webhookURL == "" {
		return nil
	}

	cfg := map[string]string{
		"url": webhookURL,
	}
	copyIfNotEmpty(cfg, "method", env["WEBHOOK_METHOD"])
	copyIfNotEmpty(cfg, "content_type", env["WEBHOOK_CONTENT_TYPE"])
	copyIfNotEmpty(cfg, "body", env["WEBHOOK_BODY"])

	if headers := normalizeQingLongWebhookHeaders(env["WEBHOOK_HEADERS"]); headers != "" {
		cfg["headers"] = headers
	}

	return cfg
}

func buildQingLongEmailConfig(env map[string]string) map[string]string {
	server := strings.TrimSpace(env["SMTP_SERVER"])
	email := strings.TrimSpace(env["SMTP_EMAIL"])
	password := strings.TrimSpace(env["SMTP_PASSWORD"])
	if server == "" || email == "" || password == "" {
		return nil
	}

	host := server
	port := "25"
	if strings.Contains(server, ":") {
		host, port, _ = strings.Cut(server, ":")
		host = strings.TrimSpace(host)
		port = strings.TrimSpace(port)
	}
	if host == "" || port == "" {
		return nil
	}

	return map[string]string{
		"smtp_host": host,
		"smtp_port": port,
		"smtp_user": email,
		"smtp_pass": password,
		"to":        email,
		"from":      email,
	}
}

func normalizeQingLongWebhookHeaders(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	if strings.HasPrefix(raw, "{") && strings.HasSuffix(raw, "}") {
		return raw
	}

	headers := map[string]string{}
	for _, line := range strings.Split(strings.ReplaceAll(raw, `\n`, "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		headers[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	if len(headers) == 0 {
		return ""
	}

	data, err := json.Marshal(headers)
	if err != nil {
		return ""
	}
	return string(data)
}

func copyIfNotEmpty(target map[string]string, key, value string) {
	value = strings.TrimSpace(value)
	if value != "" {
		target[key] = value
	}
}

func enrichManifestFromQingLongDB(dbPath string, manifest *BackupManifest) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("打开青龙数据库失败: %w", err)
	}
	defer db.Close()

	envs, err := loadQingLongEnvVars(db)
	if err != nil {
		return err
	}
	if len(envs) > 0 {
		manifest.Selection.EnvVars = true
		manifest.Data.EnvVars = append(manifest.Data.EnvVars, envs...)
	}

	tasks, err := loadQingLongTasks(db)
	if err != nil {
		return err
	}
	if len(tasks) > 0 {
		manifest.Selection.Tasks = true
		manifest.Data.Tasks = tasks
	}

	subs, err := loadQingLongSubscriptions(db)
	if err != nil {
		return err
	}
	if len(subs) > 0 {
		manifest.Selection.Subscriptions = true
		manifest.Data.Subscriptions = subs
	}

	deps, err := loadQingLongDependencies(db)
	if err != nil {
		return err
	}
	if len(deps) > 0 {
		manifest.Selection.Dependencies = true
		manifest.Data.Dependencies = deps
	}

	apps, err := loadQingLongApps(db)
	if err != nil {
		return err
	}
	if len(apps) > 0 {
		manifest.Selection.Configs = true
		manifest.Data.Configs.OpenApps = apps
	}

	return nil
}

func loadQingLongEnvVars(db *sql.DB) ([]model.EnvVar, error) {
	rows, err := db.Query("SELECT id, name, value, remarks, status, position FROM Envs ORDER BY position ASC, id ASC")
	if err != nil {
		return nil, fmt.Errorf("读取青龙环境变量失败: %w", err)
	}
	defer rows.Close()

	var result []model.EnvVar
	for rows.Next() {
		var (
			id       int
			name     string
			value    string
			remarks  sql.NullString
			status   int
			position float64
		)
		if err := rows.Scan(&id, &name, &value, &remarks, &status, &position); err != nil {
			return nil, fmt.Errorf("解析青龙环境变量失败: %w", err)
		}

		result = append(result, model.EnvVar{
			Name:      name,
			Value:     value,
			Remarks:   remarks.String,
			Enabled:   status == 0,
			Position:  position,
			SortOrder: len(result),
		})
	}

	return result, nil
}

func loadQingLongTasks(db *sql.DB) ([]model.Task, error) {
	rows, err := db.Query(`
		SELECT id, name, command, schedule, isDisabled, labels, task_before, task_after, allow_multiple_instances
		FROM Crontabs
		ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("读取青龙定时任务失败: %w", err)
	}
	defer rows.Close()

	var result []model.Task
	for rows.Next() {
		var (
			id                     int
			name                   string
			command                string
			schedule               string
			isDisabled             bool
			labelsJSON             sql.NullString
			taskBefore             sql.NullString
			taskAfter              sql.NullString
			allowMultipleInstances bool
		)
		if err := rows.Scan(&id, &name, &command, &schedule, &isDisabled, &labelsJSON, &taskBefore, &taskAfter, &allowMultipleInstances); err != nil {
			return nil, fmt.Errorf("解析青龙定时任务失败: %w", err)
		}

		schedule = strings.TrimSpace(schedule)
		if !panelcron.Parse(schedule).Valid {
			continue
		}

		task := model.Task{
			ID:                     uint(id),
			Name:                   strings.TrimSpace(name),
			Command:                strings.TrimSpace(command),
			CronExpression:         schedule,
			Status:                 model.TaskStatusEnabled,
			Timeout:                300,
			MaxRetries:             0,
			RetryInterval:          60,
			NotifyOnFailure:        true,
			NotifyOnSuccess:        false,
			AllowMultipleInstances: allowMultipleInstances,
		}
		if task.Name == "" {
			task.Name = deriveTaskNameFromCommand(task.Command)
		}
		if isDisabled {
			task.Status = model.TaskStatusDisabled
		}
		if taskBefore.Valid && strings.TrimSpace(taskBefore.String) != "" {
			value := taskBefore.String
			task.TaskBefore = &value
		}
		if taskAfter.Valid && strings.TrimSpace(taskAfter.String) != "" {
			value := taskAfter.String
			task.TaskAfter = &value
		}
		if labelsJSON.Valid {
			var labels []string
			if err := json.Unmarshal([]byte(labelsJSON.String), &labels); err == nil {
				task.SetLabelsFromSlice(labels)
			}
		}

		result = append(result, task)
	}

	return result, nil
}

func deriveTaskNameFromCommand(command string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return "导入任务"
	}
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "导入任务"
	}
	last := parts[len(parts)-1]
	base := filepath.Base(last)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func loadQingLongSubscriptions(db *sql.DB) ([]model.Subscription, error) {
	rows, err := db.Query(`
		SELECT id, name, url, schedule, type, whitelist, blacklist, dependences, branch, alias, autoAddCron, autoDelCron, is_disabled
		FROM Subscriptions
		ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("读取青龙订阅失败: %w", err)
	}
	defer rows.Close()

	var result []model.Subscription
	for rows.Next() {
		var (
			id          int
			name        string
			url         string
			schedule    string
			subType     sql.NullString
			whitelist   sql.NullString
			blacklist   sql.NullString
			dependences sql.NullString
			branch      sql.NullString
			alias       sql.NullString
			autoAdd     int
			autoDel     int
			isDisabled  bool
		)
		if err := rows.Scan(&id, &name, &url, &schedule, &subType, &whitelist, &blacklist, &dependences, &branch, &alias, &autoAdd, &autoDel, &isDisabled); err != nil {
			return nil, fmt.Errorf("解析青龙订阅失败: %w", err)
		}

		schedule = strings.TrimSpace(schedule)
		if !ValidateSubscriptionSchedule(schedule) {
			schedule = ""
		}

		result = append(result, model.Subscription{
			ID:          uint(id),
			Name:        strings.TrimSpace(name),
			Type:        inferQingLongSubscriptionType(strings.TrimSpace(url), subType.String),
			URL:         strings.TrimSpace(url),
			Branch:      strings.TrimSpace(branch.String),
			Schedule:    schedule,
			Whitelist:   strings.TrimSpace(whitelist.String),
			Blacklist:   strings.TrimSpace(blacklist.String),
			DependOn:    strings.TrimSpace(dependences.String),
			AutoAddTask: autoAdd != 0,
			AutoDelTask: autoDel != 0,
			Enabled:     !isDisabled,
			Status:      0,
			Alias:       strings.TrimSpace(alias.String),
		})
	}

	return result, nil
}

func inferQingLongSubscriptionType(url, rawType string) string {
	rawType = strings.ToLower(strings.TrimSpace(rawType))
	if strings.Contains(rawType, "file") {
		return model.SubTypeSingleFile
	}
	if strings.HasSuffix(strings.ToLower(url), ".git") {
		return model.SubTypeGitRepo
	}
	return model.SubTypeGitRepo
}

func loadQingLongDependencies(db *sql.DB) ([]BackupDependency, error) {
	rows, err := db.Query("SELECT name, type, status FROM Dependences ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("读取青龙依赖失败: %w", err)
	}
	defer rows.Close()

	var result []BackupDependency
	seen := map[string]struct{}{}
	for rows.Next() {
		var (
			name    string
			depType int
			status  int
		)
		if err := rows.Scan(&name, &depType, &status); err != nil {
			return nil, fmt.Errorf("解析青龙依赖失败: %w", err)
		}
		if strings.TrimSpace(name) == "" || status != 1 {
			continue
		}
		mappedType := mapQingLongDependencyType(depType)
		if mappedType == "" {
			continue
		}
		key := mappedType + "::" + strings.ToLower(strings.TrimSpace(name))
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, BackupDependency{
			Type: mappedType,
			Name: strings.TrimSpace(name),
		})
	}

	return result, nil
}

func mapQingLongDependencyType(depType int) string {
	switch depType {
	case 0:
		return model.DepTypeNodeJS
	case 1:
		return model.DepTypePython
	default:
		return ""
	}
}

func loadQingLongApps(db *sql.DB) ([]BackupOpenApp, error) {
	rows, err := db.Query("SELECT id, name, client_id, client_secret, scopes FROM Apps ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("读取青龙应用失败: %w", err)
	}
	defer rows.Close()

	var result []BackupOpenApp
	for rows.Next() {
		var (
			id           int
			name         string
			clientID     string
			clientSecret string
			scopesJSON   sql.NullString
		)
		if err := rows.Scan(&id, &name, &clientID, &clientSecret, &scopesJSON); err != nil {
			return nil, fmt.Errorf("解析青龙应用失败: %w", err)
		}

		scopeList := ""
		if scopesJSON.Valid && strings.TrimSpace(scopesJSON.String) != "" {
			var scopes []string
			if err := json.Unmarshal([]byte(scopesJSON.String), &scopes); err == nil {
				scopeList = strings.Join(scopes, ",")
			}
		}

		result = append(result, BackupOpenApp{
			ID:        uint(id),
			Name:      strings.TrimSpace(name),
			AppKey:    strings.TrimSpace(clientID),
			AppSecret: strings.TrimSpace(clientSecret),
			Scopes:    scopeList,
			Enabled:   true,
			RateLimit: 100,
		})
	}

	return result, nil
}

func hasAnyQingLongScriptFiles(dataDir string) bool {
	for _, path := range []string{
		filepath.Join(dataDir, "scripts"),
		filepath.Join(dataDir, "config", "extra.sh"),
		filepath.Join(dataDir, "config", "task_before.sh"),
		filepath.Join(dataDir, "config", "task_after.sh"),
		filepath.Join(dataDir, "deps"),
	} {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

func hasAnyQingLongLogFiles(dataDir string) bool {
	for _, path := range []string{filepath.Join(dataDir, "log"), filepath.Join(dataDir, "syslog")} {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

func restoreQingLongScripts(extractedDir string) error {
	dataDir, err := resolveQingLongDataDir(extractedDir)
	if err != nil {
		return err
	}

	if err := clearDirectoryContents(config.C.Data.ScriptsDir); err != nil {
		return err
	}
	if err := copyDirectoryContents(filepath.Join(dataDir, "scripts"), config.C.Data.ScriptsDir); err != nil {
		return err
	}

	for _, hook := range []string{"extra.sh", "task_before.sh", "task_after.sh"} {
		sourcePath := filepath.Join(dataDir, "config", hook)
		if _, err := os.Stat(sourcePath); err == nil {
			if err := copyFile(sourcePath, filepath.Join(config.C.Data.ScriptsDir, hook)); err != nil {
				return err
			}
		}
	}

	qlConfigDir := filepath.Join(config.C.Data.ScriptsDir, "ql-config")
	for _, fileName := range []string{"task_before.js", "task_before.py"} {
		sourcePath := filepath.Join(dataDir, "config", fileName)
		if _, err := os.Stat(sourcePath); err == nil {
			if err := copyFile(sourcePath, filepath.Join(qlConfigDir, fileName)); err != nil {
				return err
			}
		}
	}

	depsDir := filepath.Join(dataDir, "deps")
	if entries, err := os.ReadDir(depsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			sourcePath := filepath.Join(depsDir, entry.Name())
			targetPath := filepath.Join(config.C.Data.ScriptsDir, entry.Name())
			if _, err := os.Stat(targetPath); err == nil {
				continue
			}
			if err := copyFile(sourcePath, targetPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func restoreQingLongLogs(extractedDir string) error {
	dataDir, err := resolveQingLongDataDir(extractedDir)
	if err != nil {
		return err
	}

	baseTarget := filepath.Join(config.C.Data.LogDir, "qinglong-import")
	if err := os.MkdirAll(baseTarget, 0o755); err != nil {
		return err
	}
	if err := copyDirectoryContents(filepath.Join(dataDir, "log"), filepath.Join(baseTarget, "log")); err != nil {
		return err
	}
	if err := copyDirectoryContents(filepath.Join(dataDir, "syslog"), filepath.Join(baseTarget, "syslog")); err != nil {
		return err
	}
	return nil
}
