package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"

	"daidai-panel/database"
	"daidai-panel/model"
)

func SendNotification(title, content string) {
	var channels []model.NotifyChannel
	database.DB.Where("enabled = ?", true).Find(&channels)

	for _, ch := range channels {
		go sendToChannel(ch, title, content)
	}
}

func SendNotificationToChannel(channel *model.NotifyChannel, title, content string) error {
	return sendToChannel(*channel, title, content)
}

func sendToChannel(ch model.NotifyChannel, title, content string) error {
	var cfg map[string]string
	if err := json.Unmarshal([]byte(ch.Config), &cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	switch ch.Type {
	case "webhook":
		return sendWebhook(cfg, title, content)
	case "email":
		return sendEmail(cfg, title, content)
	case "telegram":
		return sendTelegram(cfg, title, content)
	case "dingtalk":
		return sendDingtalk(cfg, title, content)
	case "wecom":
		return sendWecom(cfg, title, content)
	case "bark":
		return sendBark(cfg, title, content)
	case "pushplus":
		return sendPushplus(cfg, title, content)
	case "serverchan":
		return sendServerchan(cfg, title, content)
	case "feishu":
		return sendFeishu(cfg, title, content)
	case "gotify":
		return sendGotify(cfg, title, content)
	case "pushdeer":
		return sendPushdeer(cfg, title, content)
	case "pushme":
		return sendPushMe(cfg, title, content)
	case "chanify":
		return sendChanify(cfg, title, content)
	case "igot":
		return sendIgot(cfg, title, content)
	case "qmsg":
		return sendQmsg(cfg, title, content)
	case "pushover":
		return sendPushover(cfg, title, content)
	case "discord":
		return sendDiscord(cfg, title, content)
	case "slack":
		return sendSlack(cfg, title, content)
	case "ntfy":
		return sendNtfy(cfg, title, content)
	case "wxpusher":
		return sendWxPusher(cfg, title, content)
	case "custom":
		return sendCustomWebhook(cfg, title, content)
	default:
		return fmt.Errorf("未知的通知渠道类型: %s", ch.Type)
	}
}

func httpPost(url string, body interface{}, headers map[string]string) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := NewHTTPClient(10 * time.Second)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func sendWebhook(cfg map[string]string, title, content string) error {
	webhookURL := cfg["url"]
	if webhookURL == "" {
		return fmt.Errorf("Webhook URL 为空")
	}
	body := map[string]string{"title": title, "content": content}
	return httpPost(webhookURL, body, nil)
}

func sendEmail(cfg map[string]string, title, content string) error {
	host := cfg["smtp_host"]
	port := cfg["smtp_port"]
	user := cfg["smtp_user"]
	pass := cfg["smtp_pass"]
	to := cfg["to"]
	from := cfg["from"]
	if from == "" {
		from = user
	}

	addr := host + ":" + port
	auth := smtp.PlainAuth("", user, pass, host)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, title, content)

	return smtp.SendMail(addr, auth, from, strings.Split(to, ","), []byte(msg))
}

func sendTelegram(cfg map[string]string, title, content string) error {
	token := cfg["token"]
	chatID := cfg["chat_id"]
	if token == "" || chatID == "" {
		return fmt.Errorf("Telegram token 或 chat_id 为空")
	}
	apiHost := "https://api.telegram.org"
	if v := cfg["api_host"]; v != "" {
		apiHost = strings.TrimRight(v, "/")
	} else if v := cfg["proxy"]; v != "" {
		apiHost = strings.TrimRight(v, "/")
	}
	apiURL := fmt.Sprintf("%s/bot%s/sendMessage", apiHost, token)
	body := map[string]string{
		"chat_id":    chatID,
		"text":       fmt.Sprintf("*%s*\n%s", title, content),
		"parse_mode": "Markdown",
	}
	return httpPost(apiURL, body, nil)
}

func sendDingtalk(cfg map[string]string, title, content string) error {
	webhook := cfg["webhook"]
	if webhook == "" {
		return fmt.Errorf("钉钉 Webhook URL 为空")
	}
	if secret := cfg["secret"]; secret != "" {
		timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
		stringToSign := timestamp + "\n" + secret
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(stringToSign))
		sign := url.QueryEscape(base64.StdEncoding.EncodeToString(mac.Sum(nil)))
		sep := "&"
		if !strings.Contains(webhook, "?") {
			sep = "?"
		}
		webhook = webhook + sep + "timestamp=" + timestamp + "&sign=" + sign
	}
	body := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  fmt.Sprintf("### %s\n%s", title, content),
		},
	}
	return httpPost(webhook, body, nil)
}

func sendWecom(cfg map[string]string, title, content string) error {
	webhook := cfg["webhook"]
	body := map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": fmt.Sprintf("%s\n%s", title, content)},
	}
	return httpPost(webhook, body, nil)
}

func sendBark(cfg map[string]string, title, content string) error {
	server := cfg["server"]
	key := cfg["key"]
	if key == "" {
		return fmt.Errorf("Bark Key 为空")
	}
	if server == "" {
		server = "https://api.day.app"
	}
	apiURL := fmt.Sprintf("%s/%s", strings.TrimRight(server, "/"), key)
	body := map[string]string{
		"title": title,
		"body":  content,
	}
	if v := cfg["sound"]; v != "" {
		body["sound"] = v
	}
	if v := cfg["group"]; v != "" {
		body["group"] = v
	}
	if v := cfg["icon"]; v != "" {
		body["icon"] = v
	}
	if v := cfg["level"]; v != "" {
		body["level"] = v
	}
	if v := cfg["url"]; v != "" {
		body["url"] = v
	}
	return httpPost(apiURL, body, nil)
}

func sendPushplus(cfg map[string]string, title, content string) error {
	token := cfg["token"]
	if token == "" {
		return fmt.Errorf("PushPlus Token 为空")
	}
	apiURL := "http://www.pushplus.plus/send"
	body := map[string]string{
		"token":   token,
		"title":   title,
		"content": content,
	}
	if v := cfg["topic"]; v != "" {
		body["topic"] = v
	}
	if v := cfg["template"]; v != "" {
		body["template"] = v
	}
	return httpPost(apiURL, body, nil)
}

func sendServerchan(cfg map[string]string, title, content string) error {
	key := cfg["key"]
	apiURL := fmt.Sprintf("https://sctapi.ftqq.com/%s.send", key)
	body := map[string]string{
		"title": title,
		"desp":  content,
	}
	return httpPost(apiURL, body, nil)
}

func sendFeishu(cfg map[string]string, title, content string) error {
	webhook := cfg["webhook"]
	if webhook == "" {
		return fmt.Errorf("飞书 Webhook URL 为空")
	}
	body := map[string]interface{}{
		"msg_type": "text",
		"content":  map[string]string{"text": fmt.Sprintf("%s\n%s", title, content)},
	}
	if secret := cfg["secret"]; secret != "" {
		timestamp := time.Now().Unix()
		stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
		mac := hmac.New(sha256.New, []byte(stringToSign))
		sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
		body["timestamp"] = fmt.Sprintf("%d", timestamp)
		body["sign"] = sign
	}
	return httpPost(webhook, body, nil)
}

func sendGotify(cfg map[string]string, title, content string) error {
	server := cfg["server"]
	token := cfg["token"]
	if server == "" || token == "" {
		return fmt.Errorf("Gotify 服务器地址或 Token 为空")
	}
	apiURL := fmt.Sprintf("%s/message", strings.TrimRight(server, "/"))
	priority := 5
	if v := cfg["priority"]; v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			priority = p
		}
	}
	body := map[string]interface{}{
		"title":    title,
		"message":  content,
		"priority": priority,
	}
	return httpPost(apiURL, body, map[string]string{"X-Gotify-Key": token})
}

func sendPushdeer(cfg map[string]string, title, content string) error {
	server := cfg["server"]
	key := cfg["key"]
	if server == "" {
		server = "https://api2.pushdeer.com"
	}
	apiURL := fmt.Sprintf("%s/message/push", strings.TrimRight(server, "/"))
	body := map[string]string{
		"pushkey": key,
		"text":    title,
		"desp":    content,
	}
	return httpPost(apiURL, body, nil)
}

func sendPushMe(cfg map[string]string, title, content string) error {
	server := strings.TrimSpace(cfg["server"])
	if server == "" {
		server = "https://push.i-i.me"
	}

	pushKey := strings.TrimSpace(cfg["key"])
	if pushKey == "" {
		return fmt.Errorf("PushMe push_key 为空")
	}

	form := url.Values{}
	form.Set("push_key", pushKey)
	form.Set("title", title)
	form.Set("content", content)
	if messageType := strings.TrimSpace(cfg["message_type"]); messageType != "" {
		form.Set("type", messageType)
	}

	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(server, "/"), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := NewHTTPClient(10 * time.Second)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	responseText := strings.TrimSpace(string(body))
	if responseText != "" && responseText != "success" && !strings.HasPrefix(responseText, "{") {
		return fmt.Errorf("PushMe 返回异常: %s", responseText)
	}

	return nil
}

func sendChanify(cfg map[string]string, title, content string) error {
	server := cfg["server"]
	token := cfg["token"]
	if server == "" {
		server = "https://api.chanify.net"
	}
	apiURL := fmt.Sprintf("%s/v1/sender/%s", strings.TrimRight(server, "/"), token)
	body := map[string]string{
		"title": title,
		"text":  content,
	}
	return httpPost(apiURL, body, nil)
}

func sendIgot(cfg map[string]string, title, content string) error {
	key := cfg["key"]
	apiURL := fmt.Sprintf("https://push.hellyw.com/%s", key)
	body := map[string]string{
		"title":   title,
		"content": content,
	}
	return httpPost(apiURL, body, nil)
}

func sendQmsg(cfg map[string]string, title, content string) error {
	key := strings.TrimSpace(cfg["key"])
	if key == "" {
		return fmt.Errorf("Qmsg Key 为空")
	}

	mode := strings.ToLower(strings.TrimSpace(cfg["mode"]))
	path := "send"
	if mode == "group" {
		path = "group"
	}

	apiURL := fmt.Sprintf("https://qmsg.zendee.cn/%s/%s", path, key)
	form := url.Values{}
	form.Set("msg", fmt.Sprintf("%s\n%s", title, content))
	if qq := strings.TrimSpace(cfg["qq"]); qq != "" {
		form.Set("qq", qq)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := NewHTTPClient(10 * time.Second)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Success bool   `json:"success"`
		Reason  string `json:"reason"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("Qmsg 返回无法解析: %s", strings.TrimSpace(string(body)))
	}
	if !result.Success {
		return fmt.Errorf("Qmsg 发送失败: %s", strings.TrimSpace(result.Reason))
	}

	return nil
}

func sendPushover(cfg map[string]string, title, content string) error {
	token := cfg["token"]
	user := cfg["user"]
	apiURL := "https://api.pushover.net/1/messages.json"
	body := map[string]string{
		"token":   token,
		"user":    user,
		"title":   title,
		"message": content,
	}
	return httpPost(apiURL, body, nil)
}

func sendDiscord(cfg map[string]string, title, content string) error {
	webhook := cfg["webhook"]
	body := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       title,
				"description": content,
				"color":       3447003,
			},
		},
	}
	return httpPost(webhook, body, nil)
}

func sendSlack(cfg map[string]string, title, content string) error {
	webhook := cfg["webhook"]
	body := map[string]interface{}{
		"text": fmt.Sprintf("*%s*\n%s", title, content),
	}
	return httpPost(webhook, body, nil)
}

func sendNtfy(cfg map[string]string, title, content string) error {
	server := cfg["server"]
	topic := cfg["topic"]
	if topic == "" {
		return fmt.Errorf("ntfy Topic 为空")
	}
	if server == "" {
		server = "https://ntfy.sh"
	}
	apiURL := fmt.Sprintf("%s/%s", strings.TrimRight(server, "/"), topic)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(content))
	if err != nil {
		return err
	}
	req.Header.Set("Title", title)
	if v := cfg["priority"]; v != "" {
		req.Header.Set("Priority", v)
	}
	if v := cfg["token"]; v != "" {
		req.Header.Set("Authorization", "Bearer "+v)
	}

	client := NewHTTPClient(10 * time.Second)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func sendWxPusher(cfg map[string]string, title, content string) error {
	appToken := strings.TrimSpace(cfg["app_token"])
	if appToken == "" {
		return fmt.Errorf("WxPusher appToken 为空")
	}

	uids := splitNotificationTargets(cfg["uids"])
	topicIDs, err := splitNotificationIntTargets(cfg["topic_ids"])
	if err != nil {
		return fmt.Errorf("WxPusher Topic ID 格式错误: %w", err)
	}
	if len(uids) == 0 && len(topicIDs) == 0 {
		return fmt.Errorf("WxPusher 至少需要一个 UID 或 Topic ID")
	}

	contentType := 1
	if raw := strings.TrimSpace(cfg["content_type"]); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			contentType = parsed
		}
	}

	messageContent := fmt.Sprintf("%s\n%s", title, content)
	switch contentType {
	case 2:
		messageContent = fmt.Sprintf(
			"<h1>%s</h1><br/><div style='white-space: pre-wrap;'>%s</div>",
			html.EscapeString(title),
			html.EscapeString(content),
		)
	case 3:
		messageContent = fmt.Sprintf("## %s\n\n%s", title, content)
	}

	body := map[string]interface{}{
		"appToken":    appToken,
		"content":     messageContent,
		"summary":     title,
		"contentType": contentType,
	}
	if len(uids) > 0 {
		body["uids"] = uids
	}
	if len(topicIDs) > 0 {
		body["topicIds"] = topicIDs
	}

	apiURL := "https://wxpusher.zjiecode.com/api/send/message"
	if server := strings.TrimSpace(cfg["server"]); server != "" {
		apiURL = strings.TrimRight(server, "/")
	}

	client := NewHTTPClient(10 * time.Second)
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Success bool   `json:"success"`
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil {
		if !result.Success && result.Code != 1000 {
			return fmt.Errorf("WxPusher 发送失败: %s", strings.TrimSpace(result.Msg))
		}
	}

	return nil
}

func splitNotificationTargets(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})

	result := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field != "" {
			result = append(result, field)
		}
	}
	return result
}

func splitNotificationIntTargets(raw string) ([]int, error) {
	fields := splitNotificationTargets(raw)
	result := make([]int, 0, len(fields))
	for _, field := range fields {
		value, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("无效整数 %q", field)
		}
		result = append(result, value)
	}
	return result, nil
}

func sendCustomWebhook(cfg map[string]string, title, content string) error {
	webhookURL := cfg["url"]
	method := cfg["method"]
	if method == "" {
		method = "POST"
	}

	bodyTemplate := cfg["body"]
	if bodyTemplate == "" {
		bodyTemplate = `{"title":"{{title}}","content":"{{content}}"}`
	}
	bodyStr := strings.ReplaceAll(bodyTemplate, "{{title}}", title)
	bodyStr = strings.ReplaceAll(bodyStr, "{{content}}", content)

	req, err := http.NewRequest(method, webhookURL, strings.NewReader(bodyStr))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", cfg["content_type"])
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if headerStr := cfg["headers"]; headerStr != "" {
		var headers map[string]string
		if json.Unmarshal([]byte(headerStr), &headers) == nil {
			for k, v := range headers {
				req.Header.Set(k, v)
			}
		}
	}

	client := NewHTTPClient(10 * time.Second)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
