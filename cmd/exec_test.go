package cmd

import (
	"reflect"
	"testing"
)

func TestExecCmd_Basic(t *testing.T) {
	send, params := mockSend("execute_csharp", t)
	if _, err := execCmd([]string{"Debug.Log(1)"}, send); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if (*params)["code"] != "Debug.Log(1)" {
		t.Errorf("expected code=Debug.Log(1), got %v", (*params)["code"])
	}
}

func TestExecCmd_Usings(t *testing.T) {
	send, params := mockSend("execute_csharp", t)
	if _, err := execCmd([]string{"Foo()", "--usings", "System,UnityEngine"}, send); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"System", "UnityEngine"}
	got, ok := (*params)["usings"].([]string)
	if !ok || !reflect.DeepEqual(got, want) {
		t.Errorf("usings=%v, want %v", got, want)
	}
}

func TestExecCmd_EmptyArgs(t *testing.T) {
	send, _ := mockSend("execute_csharp", t)
	_, err := execCmd(nil, send)
	if err == nil {
		t.Error("expected error for empty args")
	}
}
