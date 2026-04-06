package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveCronForSubscriptionTaskSupportsDocstringCronFilenameHeader(t *testing.T) {
	root := t.TempDir()
	scriptPath := filepath.Join(root, "bili_task_get_cookie.py")
	content := "'''\n1 9 11 11 1 bili_task_get_cookie.py\n手动运行，查看日志\n'''\nprint('hello')\n"
	if err := os.WriteFile(scriptPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	got := resolveCronForSubscriptionTask(scriptPath, "")
	if got != "1 9 11 11 1" {
		t.Fatalf("expected cron from docstring header, got %q", got)
	}
}

func TestResolveCronForSubscriptionTaskIgnoresDocstringCronForOtherFile(t *testing.T) {
	root := t.TempDir()
	scriptPath := filepath.Join(root, "actual_task.py")
	content := "'''\n1 9 11 11 1 other_task.py\n手动运行，查看日志\n'''\nprint('hello')\n"
	if err := os.WriteFile(scriptPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	got := resolveCronForSubscriptionTask(scriptPath, "0 0 * * *")
	if got != "0 0 * * *" {
		t.Fatalf("expected fallback cron for mismatched filename, got %q", got)
	}
}

func TestResolveSubscriptionTaskNamePrefersNewEnvTitle(t *testing.T) {
	root := t.TempDir()
	scriptPath := filepath.Join(root, "main.py")
	content := "\"\"\"\nnew Env('华星电信999答题');\ncron: 1 1 1 1 1\n\"\"\"\nprint('hello')\n"
	if err := os.WriteFile(scriptPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	got := resolveSubscriptionTaskName(scriptPath, "main")
	if got != "华星电信999答题" {
		t.Fatalf("expected task name from new Env title, got %q", got)
	}
}

func TestResolveSubscriptionTaskNameFallsBackToFilenameWhenNoNewEnvTitle(t *testing.T) {
	root := t.TempDir()
	scriptPath := filepath.Join(root, "main.py")
	content := "print('hello')\n"
	if err := os.WriteFile(scriptPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	got := resolveSubscriptionTaskName(scriptPath, "main")
	if got != "main" {
		t.Fatalf("expected fallback task name, got %q", got)
	}
}
