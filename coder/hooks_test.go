package coder

import (
	"strings"
	"testing"

	"github.com/msgmate-io/go-integration-interface/integrationinterface"
)

func TestNormalizeToolCallStripsWorkspaceRootCD(t *testing.T) {
	cases := []struct {
		name      string
		tool      string
		arguments string
		want      string
	}{
		{"plain root", "coder_bash", `{"command":"cd /workspace && go build ./..."}`, `{"command":"go build ./..."}`},
		{"git toplevel", "coder_bash", `{"command":"cd $(git rev-parse --show-toplevel) && ls"}`, `{"command":"ls"}`},
		{"subdirectory descent untouched", "coder_bash", `{"command":"cd backend && go test ./..."}`, ""},
		{"no prefix untouched", "coder_bash", `{"command":"go test ./..."}`, ""},
		{"non coder_bash untouched", "coder_read", `{"command":"cd /workspace && ls"}`, ""},
		{"navigation only dropped", "coder_bash", `{"command":"cd /workspace"}`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeToolCall(tc.tool, tc.arguments); got != tc.want {
				t.Fatalf("NormalizeToolCall(%q, %q) = %q, want %q", tc.tool, tc.arguments, got, tc.want)
			}
		})
	}
}

func TestPreCallInitSeedsWorkspaceFromRepoTag(t *testing.T) {
	seeded := PreCallInit(integrationinterface.AgentInitContext{
		Tags:   []string{"repo:owner/name", "branch:feature/x"},
		Config: map[string]interface{}{},
	})
	payload, ok := seeded[SelectWorkspaceToolInit]
	if !ok {
		t.Fatalf("expected %s seeded, got %v", SelectWorkspaceToolInit, seeded)
	}
	if payload[AutoSandboxGitRepoInput] != "owner/name" {
		t.Fatalf("repo = %v, want owner/name", payload[AutoSandboxGitRepoInput])
	}
	if payload[AutoSandboxGitBranchInput] != "feature/x" {
		t.Fatalf("branch = %v, want feature/x", payload[AutoSandboxGitBranchInput])
	}
	if payload[AutoSandboxFlagInput] != true {
		t.Fatalf("auto_sandbox flag = %v, want true", payload[AutoSandboxFlagInput])
	}
}

func TestPreCallInitNoRepoIsNoop(t *testing.T) {
	if seeded := PreCallInit(integrationinterface.AgentInitContext{}); seeded != nil {
		t.Fatalf("expected no seeding without a repo, got %v", seeded)
	}
}

func TestRepoFromContextCanonicalises(t *testing.T) {
	cases := map[string]string{
		"owner/name":                 "owner/name",
		"https://github.com/o/n.git": "o/n",
		"git@github.com:o/n.git":     "o/n",
		"":                           "",
		"justone":                    "",
	}
	for in, want := range cases {
		if got := RepoFromContext(map[string]interface{}{"repo": in}, nil); got != want {
			t.Fatalf("RepoFromContext(repo=%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTurnHooksBudgetExemptAndFraming(t *testing.T) {
	hooks := TurnHooks()
	if hooks.ToolBudgetExempt == nil || !hooks.ToolBudgetExempt(ScheduleToolName) {
		t.Fatalf("%s must be budget-exempt", ScheduleToolName)
	}
	if hooks.ToolBudgetExempt("coder_read") {
		t.Fatal("coder_read must not be budget-exempt")
	}
	module := "go: github.com/msgmate-io/private-integration: module lookup disabled by GOPRIVATE"
	framed := hooks.ToolFailureGuidance("coder_bash", module)
	if !strings.Contains(framed, "not materialized") {
		t.Fatalf("module failure must carry provisioning guidance, got %q", framed)
	}
	real := "--- FAIL: TestX (0.01s)"
	if got := hooks.ToolFailureGuidance("coder_bash", real); got != real {
		t.Fatalf("a genuine failure must not be reframed: got %q", got)
	}
	if got := hooks.ToolFailureGuidance("web_fetch", module); got != module {
		t.Fatalf("a non-coder tool must be untouched: got %q", got)
	}
}
