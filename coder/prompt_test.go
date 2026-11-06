package coder

import (
	"strings"
	"testing"
)

func TestComposeSystemPromptIncludesGuidance(t *testing.T) {
	prompt := ComposeSystemPrompt(map[string]interface{}{
		"system_prompt":        "You are the msgmate coder.",
		"coder_workspace_root": "/srv/checkout",
	})
	if !strings.Contains(prompt, "You are the msgmate coder.") {
		t.Fatalf("operator prompt lost: %q", prompt)
	}
	if !strings.Contains(prompt, "/srv/checkout") {
		t.Fatalf("declared workspace root not named: %q", prompt)
	}
	if !strings.Contains(prompt, GitHubTypedFirstPromptMarker) {
		t.Fatalf("typed-first GitHub block missing: %q", prompt)
	}
	if !strings.Contains(prompt, gitTypedFirstMarker) {
		t.Fatalf("typed-first git block missing: %q", prompt)
	}
	for _, tool := range GitHubTypedFirstTools() {
		if !strings.Contains(prompt, tool) {
			t.Errorf("prompt does not name %s", tool)
		}
	}
	for _, tool := range GitTypedFirstTools() {
		if !strings.Contains(prompt, tool) {
			t.Errorf("prompt does not name %s", tool)
		}
	}
}

func TestTypedFirstAppendIsIdempotent(t *testing.T) {
	once := GitHubTypedFirstAppend("You are the msgmate coder.")
	if twice := GitHubTypedFirstAppend(once); twice != once {
		t.Fatal("appending the GitHub block twice changed the prompt")
	}
	gitOnce := GitTypedFirstAppend("You are the msgmate coder.")
	if gitTwice := GitTypedFirstAppend(gitOnce); gitTwice != gitOnce {
		t.Fatal("appending the git block twice changed the prompt")
	}
}

func TestGitHubTypedFirstPurposeCoversEveryTool(t *testing.T) {
	for _, tool := range GitHubTypedFirstTools() {
		if githubTypedFirstPurpose[tool] == "" {
			t.Errorf("%s has no prompt purpose", tool)
		}
	}
}

func TestComposeSystemPromptEmptyConfigStillGuides(t *testing.T) {
	prompt := ComposeSystemPrompt(nil)
	if !strings.Contains(prompt, WorkspaceRootPromptMarker) {
		t.Fatalf("empty config should still carry workspace-root guidance: %q", prompt)
	}
}
