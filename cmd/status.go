package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/youngwoocho02/unity-cli/internal/client"
)

type UnityStatus struct {
	State        string `json:"state"`
	ProjectPath  string `json:"projectPath"`
	Port         int    `json:"port"`
	PID          int    `json:"pid"`
	UnityVersion string `json:"unityVersion"`
	Timestamp    int64  `json:"timestamp"`
}

func statusCmd(inst *client.Instance) error {
	status, err := readStatus(inst.Port)
	if err != nil {
		return fmt.Errorf("no status for port %d — Unity may not be running", inst.Port)
	}

	age := time.Since(time.UnixMilli(status.Timestamp))
	if age > 3*time.Second {
		label := "offline"
		if isBusyState(status.State) {
			label = stateLabel(status.State)
		}
		fmt.Fprintf(os.Stderr, "Unity (port %d): %s (last heartbeat %s ago)\n", status.Port, label, age.Truncate(time.Second))
		return nil
	}

	fmt.Printf("Unity (port %d): %s\n", status.Port, stateLabel(status.State))
	fmt.Printf("  Project: %s\n", status.ProjectPath)
	fmt.Printf("  Version: %s\n", status.UnityVersion)
	fmt.Printf("  PID:     %d\n", status.PID)
	return nil
}

func readStatus(port int) (*UnityStatus, error) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".unity-cli", "status", fmt.Sprintf("%d.json", port))

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var status UnityStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func checkUnityReady(port int) string {
	status, err := readStatus(port)
	if err != nil {
		return ""
	}

	age := time.Since(time.UnixMilli(status.Timestamp))
	if age > 3*time.Second {
		if isBusyState(status.State) {
			return status.State
		}
		return "offline"
	}

	return status.State
}

func isBusyState(state string) bool {
	switch state {
	case "compiling", "refreshing", "reloading", "entering_playmode":
		return true
	}
	return false
}

func waitUntilReady(port int, timeoutMs int) error {
	state := checkUnityReady(port)
	if state == "" || state == "ready" || state == "playing" || state == "paused" {
		return nil
	}

	if state == "offline" {
		return fmt.Errorf("Unity is not responding (port %d)", port)
	}

	fmt.Fprintf(os.Stderr, "Unity is %s — waiting for ready...\n", stateLabel(state))

	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		state = checkUnityReady(port)
		if state == "ready" || state == "playing" || state == "paused" {
			fmt.Fprintf(os.Stderr, "Unity is ready.\n")
			return nil
		}
		if state == "offline" || state == "" {
			return fmt.Errorf("Unity went offline while waiting")
		}
	}

	return fmt.Errorf("timed out waiting for Unity (stuck in %s)", stateLabel(state))
}

func stateLabel(state string) string {
	switch state {
	case "ready":
		return "ready"
	case "compiling":
		return "compiling scripts..."
	case "refreshing":
		return "refreshing assets..."
	case "playing":
		return "play mode"
	case "paused":
		return "play mode (paused)"
	case "reloading":
		return "reloading assemblies..."
	case "entering_playmode":
		return "entering play mode..."
	default:
		return state
	}
}
