package service

import (
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type ResourceInfo struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryTotal uint64  `json:"memory_total"`
	MemoryUsed  uint64  `json:"memory_used"`
	MemoryFree  uint64  `json:"memory_free"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskTotal   uint64  `json:"disk_total"`
	DiskUsed    uint64  `json:"disk_used"`
	DiskFree    uint64  `json:"disk_free"`
	DiskUsage   float64 `json:"disk_usage"`
	Uptime      string  `json:"uptime"`
	GoRoutines  int     `json:"goroutines"`
	GoVersion   string  `json:"go_version"`
	OS          string  `json:"os"`
	Arch        string  `json:"arch"`
	NumCPU      int     `json:"num_cpu"`
}

func GetResourceInfo() ResourceInfo {
	info := ResourceInfo{
		GoRoutines: runtime.NumGoroutine(),
		GoVersion:  runtime.Version(),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		NumCPU:     runtime.NumCPU(),
	}

	if runtime.GOOS == "linux" {
		info.MemoryTotal, info.MemoryUsed, info.MemoryFree = getLinuxMemory()
		if info.MemoryTotal > 0 {
			info.MemoryUsage = float64(info.MemoryUsed) / float64(info.MemoryTotal) * 100
		}

		info.DiskTotal, info.DiskUsed, info.DiskFree = getLinuxDisk()
		if info.DiskTotal > 0 {
			info.DiskUsage = float64(info.DiskUsed) / float64(info.DiskTotal) * 100
		}

		info.CPUUsage = getLinuxCPU()
		info.Uptime = getLinuxUptime()
	}

	return info
}

func getLinuxMemory() (total, used, free uint64) {
	out, err := exec.Command("free", "-b").Output()
	if err != nil {
		return
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return
	}
	total, _ = strconv.ParseUint(fields[1], 10, 64)
	used, _ = strconv.ParseUint(fields[2], 10, 64)
	free, _ = strconv.ParseUint(fields[3], 10, 64)
	return
}

func getLinuxDisk() (total, used, free uint64) {
	out, err := exec.Command("df", "-B1", "/").Output()
	if err != nil {
		return
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return
	}
	total, _ = strconv.ParseUint(fields[1], 10, 64)
	used, _ = strconv.ParseUint(fields[2], 10, 64)
	free, _ = strconv.ParseUint(fields[3], 10, 64)
	return
}

func getLinuxCPU() float64 {
	out, err := exec.Command("bash", "-c", `top -bn1 | grep "Cpu(s)" | awk '{print $2}'`).Output()
	if err != nil {
		return 0
	}
	val, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	return val
}

func getLinuxUptime() string {
	out, err := exec.Command("cat", "/proc/uptime").Output()
	if err != nil {
		return ""
	}
	fields := strings.Fields(string(out))
	if len(fields) < 1 {
		return ""
	}
	secs, _ := strconv.ParseFloat(fields[0], 64)
	dur := time.Duration(secs) * time.Second
	days := int(dur.Hours() / 24)
	hours := int(dur.Hours()) % 24
	mins := int(dur.Minutes()) % 60

	if days > 0 {
		return strconv.Itoa(days) + "天" + strconv.Itoa(hours) + "时" + strconv.Itoa(mins) + "分"
	}
	if hours > 0 {
		return strconv.Itoa(hours) + "时" + strconv.Itoa(mins) + "分"
	}
	return strconv.Itoa(mins) + "分"
}
