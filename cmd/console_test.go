package cmd

import (
	"reflect"
	"testing"

	"github.com/youngwoocho02/unity-cli/internal/client"
)

func mockSend(wantCmd string, t *testing.T) (sendFn, *map[string]interface{}) {
	t.Helper()
	captured := map[string]interface{}{}
	fn := func(cmd string, params interface{}) (*client.CommandResponse, error) {
		if cmd != wantCmd {
			t.Errorf("send called with command %q, want %q", cmd, wantCmd)
		}
		if p, ok := params.(map[string]interface{}); ok {
			for k, v := range p {
				captured[k] = v
			}
		}
		return &client.CommandResponse{Success: true}, nil
	}
	return fn, &captured
}

func TestConsoleCmd_Clear(t *testing.T) {
	send, params := mockSend("read_console", t)
	if _, err := consoleCmd([]string{"--clear"}, send); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if (*params)["action"] != "clear" {
		t.Errorf("expected action=clear, got %v", (*params)["action"])
	}
}

func TestConsoleCmd_Lines(t *testing.T) {
	send, params := mockSend("read_console", t)
	if _, err := consoleCmd([]string{"--lines", "50"}, send); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if (*params)["count"] != 50 {
		t.Errorf("expected count=50, got %v", (*params)["count"])
	}
}

func TestConsoleCmd_FilterMapping(t *testing.T) {
	tests := []struct {
		filter string
		want   []string
	}{
		{"error", []string{"error"}},
		{"warn", []string{"warning"}},
		{"warning", []string{"warning"}},
		{"log", []string{"log"}},
		{"all", []string{"error", "warning", "log"}},
		{"unknown", []string{"error", "warning", "log"}},
	}
	for _, tt := range tests {
		t.Run(tt.filter, func(t *testing.T) {
			send, params := mockSend("read_console", t)
			if _, err := consoleCmd([]string{"--filter", tt.filter}, send); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, ok := (*params)["types"].([]string)
			if !ok || !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filter=%q: types=%v, want %v", tt.filter, got, tt.want)
			}
		})
	}
}

func TestConsoleCmd_FilterTextFallback(t *testing.T) {
	send, params := mockSend("read_console", t)
	if _, err := consoleCmd([]string{"--filter", "NullRef"}, send); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if (*params)["filterText"] != "NullRef" {
		t.Errorf("expected filterText=NullRef, got %v", (*params)["filterText"])
	}
}

func TestConsoleCmd_Stacktrace(t *testing.T) {
	for _, v := range []string{"none", "short", "full"} {
		send, params := mockSend("read_console", t)
		if _, err := consoleCmd([]string{"--stacktrace", v}, send); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if (*params)["stacktrace"] != v {
			t.Errorf("stacktrace=%q: got %v", v, (*params)["stacktrace"])
		}
	}
}

func TestConsoleCmd_NoArgs(t *testing.T) {
	send, _ := mockSend("read_console", t)
	if _, err := consoleCmd(nil, send); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
