// Package coder is the shared, public coder policy library for the Open-Chat
// coding integrations. It holds the parts of the coding agent that are identical
// between the private `msgmate_coder` integration and the public community coder:
// the Host v5 Agent turn hooks (workspace-root seeding, `coder_bash` tolerance
// rewrite, sandbox-failure framing, bookkeeping-tool budget exemption).
//
// The package is deliberately dependency-free (stdlib plus the stable
// integration interface) so any integration, public or private, can import it
// without importing `backend/...`.
package coder

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/msgmate-io/go-integration-interface/integrationinterface"
)

// Public contract keys shared with the coder tools and the core's auto-sandbox
// hand-off.
const (
	// WorkspaceRootDefault is the sandbox path a workspace is cloned into unless
	// the chat config declares another one (mirrors the core
	// sandboxsession.DefaultWorkDir the session layer applies).
	WorkspaceRootDefault = "/workspace"
	// WorkspaceRootConfigKey declares a non-standard workspace root.
	WorkspaceRootConfigKey = "coder_workspace_root"
	// SelectWorkspaceToolInit is the chat-init tool entry every native coder
	// agent carries.
	SelectWorkspaceToolInit = "coder_select_workspace"
	// AutoSandboxGitRepoInput names the repository an auto-sandbox clones.
	AutoSandboxGitRepoInput = "git_repo"
	// AutoSandboxGitBranchInput names the branch an auto-sandbox checks out.
	AutoSandboxGitBranchInput = "git_branch"
	// AutoSandboxFlagInput switches the automatic per-chat sandbox on.
	AutoSandboxFlagInput = "auto_sandbox"
	// ScheduleToolName is the bookkeeping tool exempt from the tool-call budget.
	ScheduleToolName = "coder_schedule"
	// BashToolName is the single shell tool whose root navigation is normalised.
	BashToolName = "coder_bash"
	// RepoTagPrefix / BranchTagPrefix label the dispatch's repo and branch.
	RepoTagPrefix   = "repo:"
	BranchTagPrefix = "branch:"

	bashCommandArg = "command"
)

// TurnHooks returns the Host v5 Agent turn hooks a coder integration hands the
// core Agent runtime. Both the private and community coder register the same
// hooks, so their turn behaviour cannot drift apart.
func TurnHooks() integrationinterface.AgentHooks {
	return integrationinterface.AgentHooks{
		PreCallInit:         PreCallInit,
		NormalizeToolCall:   NormalizeToolCall,
		ToolFailureGuidance: FrameToolFailure,
		ToolBudgetExempt:    func(name string) bool { return name == ScheduleToolName },
	}
}

// PreCallInit seeds tool_init.coder_select_workspace from the interaction's
// repo/branch tags (or an explicit repo config key) so the agent never spends a
// turn on coder_select_workspace. The core fills only keys that are still absent,
// so an explicit selection always wins.
func PreCallInit(ctx integrationinterface.AgentInitContext) map[string]map[string]interface{} {
	payload := map[string]interface{}{}
	if repo := RepoFromContext(ctx.Config, ctx.Tags); repo != "" {
		payload[AutoSandboxGitRepoInput] = repo
	}
	if branch := TagValueWithPrefix(ctx.Tags, BranchTagPrefix); branch != "" {
		payload[AutoSandboxGitBranchInput] = branch
	}
	if len(payload) == 0 {
		return nil
	}
	payload[AutoSandboxFlagInput] = true
	return map[string]map[string]interface{}{SelectWorkspaceToolInit: payload}
}

// NormalizeToolCall strips a redundant leading `cd <workspace root>` from a
// `coder_bash` command before it is dispatched, so an older habit costs no call.
// It returns "" when it makes no change (the core keeps the original arguments).
func NormalizeToolCall(name, arguments string) string {
	if name != BashToolName || arguments == "" {
		return ""
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &parsed); err != nil || parsed == nil {
		return ""
	}
	command, ok := parsed[bashCommandArg].(string)
	if !ok || command == "" {
		return ""
	}
	stripped, changed := StripRedundantWorkspaceRootCD(command, WorkspaceRootDefault)
	if !changed || stripped == "" {
		return ""
	}
	parsed[bashCommandArg] = stripped
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(parsed); err != nil {
		return ""
	}
	normalized := strings.TrimRight(buf.String(), "\n")
	if strings.TrimSpace(normalized) == "" {
		return ""
	}
	return normalized
}

// RepoFromContext resolves the repository the interaction was invoked for,
// preferring an explicit `repo:<owner/name>` tag and otherwise the per-chat
// config repo keys. The value is canonicalised to `owner/name`.
func RepoFromContext(config map[string]interface{}, tags []string) string {
	if repo := CanonicalRepo(TagValueWithPrefix(tags, RepoTagPrefix)); repo != "" {
		return repo
	}
	for _, key := range []string{"repo", "repository", "github_repo", "git_repo"} {
		value, ok := config[key].(string)
		if !ok {
			continue
		}
		if repo := CanonicalRepo(value); repo != "" {
			return repo
		}
	}
	return ""
}

// TagValueWithPrefix returns the first non-empty value of a prefixed tag
// (`repo:<value>`, `branch:<value>`, case-insensitive).
func TagValueWithPrefix(tags []string, prefix string) string {
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if len(trimmed) <= len(prefix) || !strings.EqualFold(trimmed[:len(prefix)], prefix) {
			continue
		}
		if value := strings.TrimSpace(trimmed[len(prefix):]); value != "" {
			return value
		}
	}
	return ""
}

// CanonicalRepo turns a repository reference (an `owner/name` slug, a clone URL
// or a scp-style remote) into the bare `owner/name`, or "" when it carries none.
func CanonicalRepo(value string) string {
	repo := strings.TrimSpace(value)
	if repo == "" {
		return ""
	}
	repo = strings.TrimSuffix(repo, ".git")
	if idx := strings.Index(repo, "://"); idx >= 0 {
		repo = repo[idx+len("://"):]
		if slash := strings.Index(repo, "/"); slash >= 0 {
			repo = repo[slash+1:]
		}
	} else if idx := strings.Index(repo, "@"); idx >= 0 {
		repo = repo[idx+1:]
		if colon := strings.Index(repo, ":"); colon >= 0 {
			repo = repo[colon+1:]
		}
	}
	repo = strings.Trim(repo, "/")
	parts := strings.Split(repo, "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-2] + "/" + parts[len(parts)-1]
}

// WorkspaceRootFromConfig reads the declared workspace root of a sandbox config,
// falling back to WorkspaceRootDefault when the key is absent, empty or not a
// string.
func WorkspaceRootFromConfig(config map[string]interface{}) string {
	if config != nil {
		if raw, ok := config[WorkspaceRootConfigKey].(string); ok {
			if trimmed := strings.TrimSpace(raw); trimmed != "" {
				return strings.TrimRight(trimmed, "/")
			}
		}
	}
	return WorkspaceRootDefault
}

// workspaceRootCDTargets returns the literal `cd` targets equivalent to the
// workspace root (plain/quoted spellings and the git-toplevel spelling).
func workspaceRootCDTargets(root string) []string {
	clean := strings.TrimRight(strings.TrimSpace(root), "/")
	if clean == "" {
		clean = "/"
	}
	targets := []string{clean, `"` + clean + `"`, "'" + clean + "'", "$(git rev-parse --show-toplevel)"}
	if clean != "/" {
		targets = append(targets,
			clean+"/",
			`"`+clean+`/"`,
			"'"+clean+"/'",
			"$(git rev-parse --show-toplevel)/",
		)
	}
	return targets
}

// StripRedundantWorkspaceRootCD removes a leading `cd <workspace root>` from a
// shell command when it is followed by more work. changed reports whether the
// returned command differs from the input.
func StripRedundantWorkspaceRootCD(command, root string) (string, bool) {
	trimmed := strings.TrimSpace(command)
	if trimmed == "" {
		return command, false
	}
	for _, target := range workspaceRootCDTargets(root) {
		prefix := "cd " + target
		if !strings.HasPrefix(trimmed, prefix) {
			continue
		}
		rest := strings.TrimSpace(trimmed[len(prefix):])
		if rest == "" {
			return "", true
		}
		for _, separator := range []string{"&&", ";"} {
			if strings.HasPrefix(rest, separator) {
				if next := strings.TrimSpace(rest[len(separator):]); next != "" {
					return next, true
				}
			}
		}
	}
	return command, false
}
