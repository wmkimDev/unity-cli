package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func writeStatusFile(t *testing.T, status UnityStatus) string {
	t.Helper()
	home := t.TempDir()
	statusDir := filepath.Join(home, ".unity-cli", "status")
	if err := os.MkdirAll(statusDir, 0755); err != nil {
		t.Fatalf("failed to create status dir: %v", err)
	}
	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("failed to marshal status: %v", err)
	}
	path := filepath.Join(statusDir, fmt.Sprintf("%d.json", status.Port))
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write status file: %v", err)
	}
	return home
}

func TestReadStatus_ValidFile(t *testing.T) {
	want := UnityStatus{
		State:        "ready",
		ProjectPath:  "/home/user/MyProject",
		Port:         8090,
		PID:          12345,
		UnityVersion: "6000.3.10f1",
		Timestamp:    1000000,
	}

	home := writeStatusFile(t, want)
	t.Setenv("HOME", home)

	got, err := readStatus(8090)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.State != want.State {
		t.Errorf("State: got %q, want %q", got.State, want.State)
	}
	if got.Port != want.Port {
		t.Errorf("Port: got %d, want %d", got.Port, want.Port)
	}
	if got.ProjectPath != want.ProjectPath {
		t.Errorf("ProjectPath: got %q, want %q", got.ProjectPath, want.ProjectPath)
	}
}

func TestReadStatus_MissingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	_, err := readStatus(9999)
	if err == nil {
		t.Error("expected error for missing status file")
	}
}

func TestReadStatus_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	statusDir := filepath.Join(dir, ".unity-cli", "status")
	if err := os.MkdirAll(statusDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(statusDir, "8090.json"), []byte("not json"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	t.Setenv("HOME", dir)

	_, err := readStatus(8090)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
