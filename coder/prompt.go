package coder

import (
	"sort"
	"strings"
)

// Coder system-prompt composition, shared by the private and community coder
// integrations. It used to live in the core (backend/api/msgmate +
// backend/nativecoder); the integration owns it now so it is delivered from
// inside the integration via the Host v5 Agent spec.

// WorkspaceRootPromptMarker is the leading token of the workspace-root guidance,
// used to keep the append idempotent when a caller already composed it.
const WorkspaceRootPromptMarker = "Workspace root:"

// GitHub typed-first GitHub prompt block (ci#514 phase 1).
const (
	// GitHubTypedFirstPromptMarker is the stable first line of the block.
	GitHubTypedFirstPromptMarker = "GitHub typed tools (use these instead of `gh`):"

	githubToolIssueList    = "github_issue_list"
	githubToolIssueView    = "github_issue_view"
	githubToolIssueCreate  = "github_issue_create"
	githubToolIssueComment = "github_issue_comment"
	githubToolPRList       = "github_pr_list"
	githubToolPRView       = "github_pr_view"
	githubToolPRCreate     = "github_pr_create"
	githubToolPRComment    = "github_pr_comment"
	githubToolPRReview     = "github_pr_review"
	githubToolPRMerge      = "github_pr_merge"
	githubToolAuthStatus   = "github_auth_status"
	githubToolAPI          = "github_api"
	githubToolRunList      = "github_run_list"
	githubToolRunView      = "github_run_view"
	githubToolAwait        = "github_await"
)

// githubTypedFirstPurpose is the one-line purpose of every typed GitHub tool as
// the coder system prompt presents it.
var githubTypedFirstPurpose = map[string]string{
	githubToolIssueList:    "list and filter issues (`gh issue list`)",
	githubToolIssueView:    "read one issue with its comments (`gh issue view`)",
	githubToolIssueCreate:  "open an issue (`gh issue create`), mutating",
	githubToolIssueComment: "comment on an issue (`gh issue comment`), mutating",
	githubToolPRList:       "list and filter pull requests (`gh pr list`)",
	githubToolPRView:       "read one pull request, its diff and checks (`gh pr view`, `gh pr diff`)",
	githubToolPRCreate:     "open a pull request (`gh pr create`), mutating",
	githubToolPRComment:    "comment on a pull request (`gh pr comment`), mutating",
	githubToolPRReview:     "submit a review (`gh pr review`), mutating",
	githubToolPRMerge:      "merge a pull request (`gh pr merge`), admin",
	githubToolAuthStatus:   "check the GitHub auth and token scope (`gh auth status`)",
	githubToolAPI:          "read-only REST/GraphQL call (`gh api`)",
	githubToolRunList:      "list workflow runs (`gh run list`), read-only",
	githubToolRunView:      "read one workflow run and its jobs (`gh run view`), read-only",
	githubToolAwait:        "block until a workflow run or PR checks settle, then return structured terminal state, never busy-wait (`gh pr checks --watch` / `gh run watch` / poll loops)",
}

// githubTypedFirstVerifiedWrites names the typed GitHub tools whose writes are
// re-read with a typed read tool and reported as verified (rather than claimed).
// It is the static form of the shipped write-verification plans.
var githubTypedFirstVerifiedWrites = []string{
	githubToolIssueComment,
	githubToolIssueCreate,
	githubToolPRComment,
	githubToolPRCreate,
	githubToolPRMerge,
	githubToolPRReview,
}

// GitHubTypedFirstTools returns the typed GitHub tool names in stable order.
func GitHubTypedFirstTools() []string {
	tools := make([]string, 0, len(githubTypedFirstPurpose))
	for tool := range githubTypedFirstPurpose {
		tools = append(tools, tool)
	}
	sort.Strings(tools)
	return tools
}

// GitHubTypedFirstGuidance is the typed-first GitHub prompt block.
func GitHubTypedFirstGuidance() string {
	lines := []string{
		GitHubTypedFirstPromptMarker,
		"- use them instead of `gh ...` inside coder_bash: structured results, fewer confirmation round-trips, auditable calls.",
	}
	for _, tool := range GitHubTypedFirstTools() {
		lines = append(lines, "- "+tool+": "+githubTypedFirstPurpose[tool])
	}
	verified := append([]string(nil), githubTypedFirstVerifiedWrites...)
	sort.Strings(verified)
	lines = append(lines,
		"- a mutating GitHub write is re-read with its typed read tool and reported as verified, never claimed: "+strings.Join(verified, ", "),
		"- fall back to `gh` in coder_bash only for a workflow no typed tool covers, and say which one it is.",
	)
	return strings.Join(lines, "\n")
}

// GitHubTypedFirstAppend returns prompt with the typed-first GitHub block
// appended. It is idempotent and never rewrites the operator prompt.
func GitHubTypedFirstAppend(prompt string) string {
	block := GitHubTypedFirstGuidance()
	if block == "" || strings.Contains(prompt, GitHubTypedFirstPromptMarker) {
		return prompt
	}
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return block
	}
	return trimmed + "\n\n" + block
}

// Git typed-first git prompt block (ci#514 phase 3).
const (
	gitTypedFirstMarker = "Prefer the typed git tools over coder_bash for every git workflow:"

	gitShellToolStatus   = "coder_git_status"
	gitShellToolLog      = "coder_git_log"
	gitShellToolShow     = "coder_git_show"
	gitShellToolDiff     = "coder_git_diff"
	gitShellToolBlame    = "coder_git_blame"
	gitShellToolRemote   = "coder_git_remote"
	gitShellToolBranch   = "coder_git_branch"
	gitShellToolCheckout = "coder_git_checkout"
	gitShellToolAdd      = "coder_git_add"
	gitShellToolCommit   = "coder_git_commit"
	gitShellToolPush     = "coder_git_push"
	gitShellToolFetch    = "coder_git_fetch"
	gitShellToolStash    = "coder_git_stash"
	gitShellToolTag      = "coder_git_tag"

	gitShellDefaultBranch = "main"
)

func gitTypedFirstReadTools() []string {
	return []string{gitShellToolStatus, gitShellToolLog, gitShellToolShow, gitShellToolDiff, gitShellToolBlame, gitShellToolRemote}
}

func gitTypedFirstWriteTools() []string {
	return []string{gitShellToolBranch, gitShellToolCheckout, gitShellToolAdd, gitShellToolCommit, gitShellToolPush, gitShellToolFetch, gitShellToolStash, gitShellToolTag}
}

// GitTypedFirstTools is the flat list of every typed git tool the block names.
func GitTypedFirstTools() []string {
	tools := append([]string(nil), gitTypedFirstReadTools()...)
	return append(tools, gitTypedFirstWriteTools()...)
}

// GitTypedFirstGuidance is the typed-first git prompt block.
func GitTypedFirstGuidance() string {
	lines := []string{
		gitTypedFirstMarker,
		"  - reads: " + strings.Join(gitTypedFirstReadTools(), ", "),
		"  - writes: " + strings.Join(gitTypedFirstWriteTools(), ", "),
		"The typed surface refuses to push " + gitShellDefaultBranch + " and refuses to create an empty commit; never work around that with coder_bash.",
	}
	return strings.Join(lines, "\n")
}

// GitTypedFirstAppend returns systemPrompt with the typed-first git guidance
// appended. It is idempotent.
func GitTypedFirstAppend(systemPrompt string) string {
	if strings.Contains(systemPrompt, gitTypedFirstMarker) {
		return systemPrompt
	}
	guidance := GitTypedFirstGuidance()
	trimmed := strings.TrimSpace(systemPrompt)
	if trimmed == "" {
		return guidance
	}
	return trimmed + "\n\n" + guidance
}

// ComposeSystemPrompt composes the coder system prompt from a chat's shared
// config: the configured `system_prompt`, then the workspace-root, GitHub
// typed-first and git typed-first guidance. It is idempotent and mirrors the
// core's previous composition exactly.
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

// WorkspaceRootGuidanceLinesFor returns the model-facing prompt lines that
// state the workspace root up front.
func WorkspaceRootGuidanceLinesFor(root string) []string {
	resolved := strings.TrimRight(strings.TrimSpace(root), "/")
	if resolved == "" {
		resolved = WorkspaceRootDefault
	}
	return []string{
		"Workspace root: `" + resolved + "` is the root of your checkout. Every `coder_bash` command already " +
			"starts there, and a workspace-relative path in `coder_read` / `coder_list` / `coder_glob` / " +
			"`coder_grep` resolves against it.",
		"Do not prefix a command with `cd " + resolved + "` or `cd $(git rev-parse --show-toplevel)`: it is a " +
			"no-op that wastes a turn. Pass the workspace-relative path (eg `go test ./backend/nativecoder/...`) " +
			"and use `cd <subdir> && ...` only to descend into a subdirectory.",
	}
}

// WorkspaceRootGuidanceFor is the workspace-root guidance for a declared root.
func WorkspaceRootGuidanceFor(root string) string {
	return strings.Join(WorkspaceRootGuidanceLinesFor(root), " ")
}
