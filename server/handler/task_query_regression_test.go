package handler_test

import (
	"net/http"
	"strconv"
	"testing"

	"daidai-panel/database"
	"daidai-panel/model"
	"daidai-panel/testutil"
)

func TestTaskListPlacesDisabledTasksAfterActiveOnes(t *testing.T) {
	testutil.SetupTestEnv(t)

	engine := newProtectedRouter()
	user := testutil.MustCreateUser(t, "operator", "operator")
	accessToken := testutil.MustCreateAccessToken(t, user.Username, user.Role)

	tasks := []*model.Task{
		{Name: "disabled pinned", Command: "echo disabled-pinned", CronExpression: "0 0 * * *", IsPinned: true},
		{Name: "enabled pinned", Command: "echo enabled-pinned", CronExpression: "0 0 * * *", IsPinned: true},
		{Name: "enabled normal", Command: "echo enabled-normal", CronExpression: "0 0 * * *"},
		{Name: "disabled normal", Command: "echo disabled-normal", CronExpression: "0 0 * * *"},
	}
	for _, task := range tasks {
		if err := database.DB.Create(task).Error; err != nil {
			t.Fatalf("create task %q: %v", task.Name, err)
		}
	}
	if err := database.DB.Model(tasks[0]).Update("status", model.TaskStatusDisabled).Error; err != nil {
		t.Fatalf("set disabled status for %q: %v", tasks[0].Name, err)
	}
	if err := database.DB.Model(tasks[1]).Update("status", model.TaskStatusEnabled).Error; err != nil {
		t.Fatalf("set enabled status for %q: %v", tasks[1].Name, err)
	}
	if err := database.DB.Model(tasks[2]).Update("status", model.TaskStatusEnabled).Error; err != nil {
		t.Fatalf("set enabled status for %q: %v", tasks[2].Name, err)
	}
	if err := database.DB.Model(tasks[3]).Update("status", model.TaskStatusDisabled).Error; err != nil {
		t.Fatalf("set disabled status for %q: %v", tasks[3].Name, err)
	}

	rec := performRequest(engine, http.MethodGet, "/api/v1/tasks", map[string]string{
		"Authorization": "Bearer " + accessToken,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	payload := decodeJSONMap(t, rec)
	items, ok := payload["data"].([]interface{})
	if !ok {
		t.Fatalf("expected data array, got %#v", payload["data"])
	}
	if len(items) < 4 {
		t.Fatalf("expected at least 4 tasks, got %d", len(items))
	}

	gotNames := make([]string, 0, 4)
	for i := 0; i < 4; i++ {
		item, ok := items[i].(map[string]interface{})
		if !ok {
			t.Fatalf("expected task object at %d, got %#v", i, items[i])
		}
		gotNames = append(gotNames, item["name"].(string))
	}

	wantNames := []string{
		"enabled pinned",
		"enabled normal",
		"disabled pinned",
		"disabled normal",
	}
	for i, want := range wantNames {
		if gotNames[i] != want {
			t.Fatalf("expected order %v, got %v", wantNames, gotNames)
		}
	}
}

func TestTaskListMapsSubscriptionLabelsToSubscriptionNames(t *testing.T) {
	testutil.SetupTestEnv(t)

	engine := newProtectedRouter()
	user := testutil.MustCreateUser(t, "operator", "operator")
	accessToken := testutil.MustCreateAccessToken(t, user.Username, user.Role)

	subscription := &model.Subscription{
		Name:    "kele",
		Type:    model.SubTypeGitRepo,
		URL:     "https://github.com/Aellyt/kele.git",
		Enabled: true,
	}
	if err := database.DB.Create(subscription).Error; err != nil {
		t.Fatalf("create subscription: %v", err)
	}

	task := &model.Task{
		Name:           "subscription task",
		Command:        "task kele/main.js",
		CronExpression: "0 0 * * *",
		Status:         model.TaskStatusEnabled,
	}
	task.SetLabelsFromSlice([]string{"manual", "subscription:" + strconv.FormatUint(uint64(subscription.ID), 10)})
	if err := database.DB.Create(task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}

	rec := performRequest(engine, http.MethodGet, "/api/v1/tasks", map[string]string{
		"Authorization": "Bearer " + accessToken,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	payload := decodeJSONMap(t, rec)
	items, ok := payload["data"].([]interface{})
	if !ok || len(items) == 0 {
		t.Fatalf("expected non-empty data array, got %#v", payload["data"])
	}

	firstItem, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected task object, got %#v", items[0])
	}

	displayLabels, ok := firstItem["display_labels"].([]interface{})
	if !ok {
		t.Fatalf("expected display_labels array, got %#v", firstItem["display_labels"])
	}

	got := make([]string, 0, len(displayLabels))
	for _, item := range displayLabels {
		got = append(got, item.(string))
	}

	expected := []string{"manual", "kele"}
	for _, want := range expected {
		found := false
		for _, label := range got {
			if label == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected display_labels to contain %q, got %v", want, got)
		}
	}
}
