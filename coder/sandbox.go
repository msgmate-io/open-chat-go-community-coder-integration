package coder

import "strings"

// Sandbox-readiness policy for failed *coder* tool calls (ci#739). The
// integration owns the sandbox, so it owns the guidance injected when a coder
// tool fails because the sandbox was not provisioned (toolchain not on PATH,
// integration modules never materialized).

const (
	sandboxGoBinDir            = "/usr/local/go/bin"
	sandboxPrivateModuleOrg    = "github.com/msgmate-io/*"
	sandboxProfileSetupCommand = "development/scripts/build_tools.sh >/dev/null && $(development/scripts/build_tools.sh) prepare --profile full-ci"
	sandboxToolPrefix          = "coder_"
)

var sandboxToolchainPathEntries = []string{sandboxGoBinDir, "/dev_bin", "/go/bin"}

var sandboxToolchainTools = []string{"go", "gofmt"}

var sandboxShellNotFoundMarkers = []string{
	": command not found",
	": not found",
	"not found in $path",
	"no such file or directory",
}

var sandboxPrivateModuleMarkers = []string{
	"module lookup disabled",
	"no required module provides package",
	"cannot find module providing package",
	"missing go.sum entry",
	"unrecognized import path",
	"could not read username for",
	"terminal prompts disabled",
	"repository not found",
	"public key authentication failed",
}

// FrameToolFailure appends sandbox-readiness guidance to a failed coder tool
// result, or returns it unchanged when there is nothing to say.
func FrameToolFailure(toolName, toolResult string) string {
	if !strings.HasPrefix(strings.TrimSpace(toolName), sandboxToolPrefix) {
		return toolResult
	}
	if hint, ok := sandboxPrivateModuleHint(toolResult); ok {
		return toolResult + " " + hint
	}
	if hint, ok := sandboxToolMissingHintFromOutput(toolResult); ok {
		return toolResult + " " + hint
	}
	return toolResult
}

func sandboxPrivateModuleHint(output string) (string, bool) {
	lowered := strings.ToLower(output)
	for _, marker := range sandboxPrivateModuleMarkers {
		if !strings.Contains(lowered, marker) {
			continue
		}
		return "The integration modules are not materialized in this sandbox (not a source defect). Run `" +
			sandboxProfileSetupCommand + "` once, then re-run the check. The Go workspace resolves " +
			"integrations from the profile checkouts, not from the module proxy; PR CI remains the " +
			"authoritative full-suite gate, so a locally unresolvable private module alone must not " +
			"block the change.", true
	}
	return "", false
}

func sandboxToolMissingHint(tool string) (string, bool) {
	name := strings.TrimSpace(tool)
	for _, known := range sandboxToolchainTools {
		if name != known {
			continue
		}
		return "`" + known + "` is not on PATH. The sandbox image installs the toolchain at " +
			sandboxGoBinDir + "; call it as " + sandboxGoBinDir + "/" + known +
			" or prepend " + strings.Join(sandboxToolchainPathEntries, ":") + " to PATH before giving up on local checks.", true
	}
	return "", false
}

func sandboxToolMissingHintFromOutput(output string) (string, bool) {
	lowered := strings.ToLower(output)
	for _, tool := range sandboxToolchainTools {
		for _, marker := range sandboxShellNotFoundMarkers {
			if strings.Contains(lowered, tool+marker) {
				return sandboxToolMissingHint(tool)
			}
		}
	}
	return "", false
}
