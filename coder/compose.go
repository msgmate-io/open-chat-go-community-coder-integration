package coder

import "strings"

// WorkspaceRootPromptMarker is the leading token of the workspace-root guidance,
// used to keep the append idempotent when a caller already composed it.
const WorkspaceRootPromptMarker = "Workspace root:"

// ComposeSystemPrompt composes the coder system prompt from a chat's shared
// config: the configured `system_prompt`, then the workspace-root, GitHub
// typed-first and git typed-first guidance. It is idempotent and mirrors the
// core's previous composition.
func ComposeSystemPrompt(config map[string]interface{}) string {
	systemPrompt := ""
	if config != nil {
		if raw, ok := config["system_prompt"].(string); ok {
			systemPrompt = raw
		}
	}
	root := WorkspaceRootFromConfig(config)
	guidance := strings.TrimSpace(WorkspaceRootGuidanceFor(root))
	guidance = GitHubTypedFirstAppend(guidance)
	guidance = GitTypedFirstAppend(guidance)
	if guidance == "" {
		return systemPrompt
	}
	trimmed := strings.TrimSpace(systemPrompt)
	if strings.Contains(trimmed, WorkspaceRootPromptMarker) {
		return GitTypedFirstAppend(GitHubTypedFirstAppend(systemPrompt))
	}
	if trimmed == "" {
		return guidance
	}
	return trimmed + "\n\n" + guidance
}
