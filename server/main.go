package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"daidai-panel/config"
	"daidai-panel/database"
	"daidai-panel/middleware"
	"daidai-panel/model"
	"daidai-panel/router"
	"daidai-panel/service"

	"github.com/gin-gonic/gin"
)

type startupLogFilter struct {
	dst io.Writer
}

func (w *startupLogFilter) Write(p []byte) (int, error) {
	text := string(p)
	if shouldSuppressStartupLog(text) {
		return len(p), nil
	}
	return w.dst.Write(p)
}

func shouldSuppressStartupLog(text string) bool {
	markers := []string{
		"database connected:",
		"added missing column:",
		"column check completed",
		"scheduler v2 started:",
		"scheduler v2 initialized with",
		"subscription scheduler initialized with",
		"resource watcher started",
		"server starting on",
	}

	for _, marker := range markers {
		if strings.Contains(text, marker) {
			return true
		}
	}

	return false
}

func buildAccessURLs(port int) []string {
	if port <= 0 {
		return nil
	}

	seen := map[string]struct{}{}
	var urls []string

	addURL := func(host string) {
		host = strings.TrimSpace(host)
		if host == "" {
			return
		}
		url := fmt.Sprintf("http://%s:%d", host, port)
		if _, exists := seen[url]; exists {
			return
		}
		seen[url] = struct{}{}
		urls = append(urls, url)
	}

	addURL("127.0.0.1")
	addURL("localhost")

	ifaces, err := net.Interfaces()
	if err != nil {
		return urls
	}

	var localIPs []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			ip = ip.To4()
			if ip == nil || ip.IsLoopback() {
				continue
			}

			localIPs = append(localIPs, ip.String())
		}
	}

	sort.Strings(localIPs)
	for _, ip := range localIPs {
		addURL(ip)
	}

	return urls
}

func printStartupSummary(port int) {
	urls := buildAccessURLs(port)
	fmt.Println("呆呆面板已经启动")
	if len(urls) == 0 {
		fmt.Printf("访问地址：http://127.0.0.1:%d\n", port)
		return
	}

	fmt.Println("访问地址：")
	for _, url := range urls {
		fmt.Println(url)
	}
	fmt.Printf("请使用上面显示的宿主机访问地址，不要直接使用容器内端口 %d/%d。\n", 5700, 5701)
}

func setupPanelLog(dataDir string) io.Writer {
	logFilePath := filepath.Join(dataDir, "panel.log")
	if err := os.MkdirAll(filepath.Dir(logFilePath), 0o755); err != nil {
		return os.Stdout
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return os.Stdout
	}

	return io.MultiWriter(os.Stdout, logFile)
}

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	panelWriter := setupPanelLog(cfg.Data.Dir)
	log.SetOutput(&startupLogFilter{dst: panelWriter})
	gin.DefaultWriter = panelWriter
	gin.DefaultErrorWriter = panelWriter

	database.Init(&cfg.Database)

	database.AutoMigrate(
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
	)

	database.EnsureColumns()

	model.InitDefaultConfigs()
	if err := middleware.ConfigureTrustedProxyCIDRs(model.GetRegisteredConfig("trusted_proxy_cidrs")); err != nil {
		log.Fatalf("failed to configure trusted proxies: %v", err)
	}

	verifyInstalledDeps()

	service.InitSchedulerV2()
	defer service.ShutdownSchedulerV2()

	service.InitSubscriptionScheduler()
	defer service.ShutdownSubscriptionScheduler()

	service.StartResourceWatcher()
	defer service.StopResourceWatcher()

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	if err := engine.SetTrustedProxies(middleware.CurrentTrustedProxyCIDRs()); err != nil {
		log.Fatalf("failed to apply trusted proxies to gin engine: %v", err)
	}
	engine.RemoteIPHeaders = []string{"X-Real-IP", "X-Forwarded-For"}
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	router.Setup(engine)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("server failed: %v", err)
	}

	log.SetOutput(panelWriter)
	printStartupSummary(cfg.Server.Port)

	if err := engine.RunListener(listener); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func verifyInstalledDeps() {
	var deps []model.Dependency
	database.DB.Where("status = ?", model.DepStatusInstalled).Find(&deps)
	if len(deps) == 0 {
		return
	}

	depsDir := filepath.Join(config.C.Data.Dir, "deps")
	resetCount := 0

	for _, dep := range deps {
		exists := false
		switch dep.Type {
		case model.DepTypeNodeJS:
			modDir := filepath.Join(depsDir, "nodejs", "node_modules", dep.Name)
			if _, err := os.Stat(modDir); err == nil {
				exists = true
			}
		case model.DepTypePython:
			venvPip := filepath.Join(depsDir, "python", "venv", "bin", "pip")
			if _, err := os.Stat(venvPip); err == nil {
				out, err := exec.Command(venvPip, "show", dep.Name).CombinedOutput()
				if err == nil && strings.Contains(string(out), "Name:") {
					exists = true
				}
			}
		case model.DepTypeLinux:
			if out, err := exec.Command("which", dep.Name).CombinedOutput(); err == nil && len(strings.TrimSpace(string(out))) > 0 {
				exists = true
			} else if exec.Command("apk", "info", "-e", dep.Name).Run() == nil {
				exists = true
			}
		}

		if !exists {
			database.DB.Model(&dep).Updates(map[string]interface{}{
				"status": model.DepStatusFailed,
				"log":    dep.Log + "\n[启动校验] 依赖未检测到，可能因容器重建而丢失，请重新安装",
			})
			resetCount++
			log.Printf("dep verify: %s/%s not found, status reset to failed", dep.Type, dep.Name)
		}
	}

	if resetCount > 0 {
		log.Printf("dep verify: %d dependencies reset to failed (not found on system)", resetCount)
	}

	var stale []model.Dependency
	database.DB.Where("status IN ?", []string{model.DepStatusInstalling, model.DepStatusRemoving}).Find(&stale)
	for _, dep := range stale {
		database.DB.Model(&dep).Updates(map[string]interface{}{
			"status": model.DepStatusFailed,
			"log":    dep.Log + "\n[启动校验] 操作因服务重启而中断",
		})
		log.Printf("dep verify: %s/%s was %s, reset to failed", dep.Type, dep.Name, dep.Status)
	}
}
