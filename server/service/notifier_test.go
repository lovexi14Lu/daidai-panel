package service

import (
	"reflect"
	"testing"
)

func TestSplitNotificationTargets(t *testing.T) {
	got := splitNotificationTargets("uid-a; uid-b,\nuid-c\tuid-d")
	want := []string{"uid-a", "uid-b", "uid-c", "uid-d"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected targets: got %v want %v", got, want)
	}
}

func TestSplitNotificationIntTargets(t *testing.T) {
	got, err := splitNotificationIntTargets("101; 102,103")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []int{101, 102, 103}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected int targets: got %v want %v", got, want)
	}
}

func TestSplitNotificationIntTargetsRejectsInvalidValue(t *testing.T) {
	if _, err := splitNotificationIntTargets("101;abc"); err == nil {
		t.Fatal("expected invalid topic id to return an error")
	}
}
