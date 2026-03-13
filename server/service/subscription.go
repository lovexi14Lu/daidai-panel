package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"daidai-panel/config"
	"daidai-panel/database"
	"daidai-panel/model"
)

func PullSubscription(sub *model.Subscription) (string, error) {
	startTime := time.Now()

	var sshKeyPath string
	if sub.SSHKeyID != nil {
		var sshKey model.SSHKey
		if err := database.DB.First(&sshKey, *sub.SSHKeyID).Error; err == nil {
			tmpFile, err := writeTempSSHKey(sshKey.PrivateKey)
			if err != nil {
				return "", fmt.Errorf("写入 SSH 密钥失败: %w", err)
			}
			defer os.Remove(tmpFile)
			sshKeyPath = tmpFile
		}
	}

	var output string
	var pullErr error

	switch sub.Type {
	case model.SubTypeSingleFile:
		output, pullErr = pullSingleFile(sub, sshKeyPath)
	default:
		output, pullErr = pullGitRepo(sub, sshKeyPath)
	}

	duration := time.Since(startTime).Seconds()

	status := 0
	content := output
	if pullErr != nil {
		status = 1
		content = fmt.Sprintf("%s\nError: %s", output, pullErr.Error())
	}

	subLog := model.SubLog{
		SubscriptionID: sub.ID,
		Status:         status,
		Content:        content,
		Duration:        duration,
	}
	database.DB.Create(&subLog)

	now := time.Now()
	database.DB.Model(sub).Updates(map[string]interface{}{
		"last_pull_at": &now,
		"status":       status,
	})

	return output, pullErr
}

func pullGitRepo(sub *model.Subscription, sshKeyPath string) (string, error) {
	saveDir := sub.SaveDir
	if saveDir == "" {
		saveDir = sub.Alias
		if saveDir == "" {
			parts := strings.Split(sub.URL, "/")
			saveDir = strings.TrimSuffix(parts[len(parts)-1], ".git")
		}
	}

	destDir := filepath.Join(config.C.Data.ScriptsDir, saveDir)

	if IsGitRepo(destDir) {
		GitReset(destDir)
		output, err := GitPull(destDir, sshKeyPath)
		return output, err
	}

	output, err := GitClone(sub.URL, sub.Branch, destDir, sshKeyPath)
	return output, err
}

func pullSingleFile(sub *model.Subscription, _ string) (string, error) {
	saveDir := sub.SaveDir
	if saveDir == "" {
		saveDir = "downloads"
	}

	parts := strings.Split(sub.URL, "/")
	filename := parts[len(parts)-1]
	if sub.Alias != "" {
		filename = sub.Alias
	}

	destPath := filepath.Join(config.C.Data.ScriptsDir, saveDir, filename)
	output, err := DownloadFile(sub.URL, destPath)
	return output, err
}

func writeTempSSHKey(privateKey string) (string, error) {
	tmpFile, err := os.CreateTemp("", "ssh_key_*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(privateKey); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	os.Chmod(tmpFile.Name(), 0600)
	return tmpFile.Name(), nil
}
