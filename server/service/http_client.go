package service

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"daidai-panel/model"
)

func NewHTTPClient(timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	proxyURL := strings.TrimSpace(model.GetRegisteredConfig("proxy_url"))
	if proxyURL != "" {
		if parsed, err := url.Parse(proxyURL); err == nil {
			transport.Proxy = http.ProxyURL(parsed)
		}
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

func AppendProxyEnv(env []string) []string {
	proxyURL := strings.TrimSpace(model.GetRegisteredConfig("proxy_url"))
	if proxyURL == "" {
		return env
	}

	keys := []string{
		"HTTP_PROXY",
		"HTTPS_PROXY",
		"ALL_PROXY",
		"http_proxy",
		"https_proxy",
		"all_proxy",
	}

	for _, key := range keys {
		env = append(env, key+"="+proxyURL)
	}
	return env
}
