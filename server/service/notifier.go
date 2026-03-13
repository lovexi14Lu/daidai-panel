package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	case "chanify":
		return sendChanify(cfg, title, content)
	case "igot":
		return sendIgot(cfg, title, content)
	case "pushover":
		return sendPushover(cfg, title, content)
	case "discord":
		return sendDiscord(cfg, title, content)
	case "slack":
		return sendSlack(cfg, title, content)
	case "ntfy":
		return sendNtfy(cfg, title, content)
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

	client := &http.Client{Timeout: 10 * time.Second}
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

	client := &http.Client{Timeout: 10 * time.Second}
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

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
