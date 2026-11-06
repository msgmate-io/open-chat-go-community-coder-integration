package coder

// This file is the integration-facing API the core consumes through the
// integrationinterface Definition hooks (AgentConfig / AgentDefaults /
// AgentInteractionDefaults / AgentInteractionTags / AgentChatSetup). It composes
// the coder policy that used to live in the core, so the coding feature is
// delivered from inside the coder integration.

// AgentConfig reports whether a shared config declares a coder agent.
func AgentConfig(config map[string]interface{}) bool {
	return IsCoderAgentConfig(config)
}

// BotDefaults merges the bot-level coder defaults into a bot's shared config:
// the default-ON auto-consolidation block, the dev-agent hive-mind dispatch
// family and the autonomous-confirmation contract. It deliberately does NOT
// record the tool-call budget snapshot: that belongs to a chat, and the core
// applies it where a chat's config is built.
func BotDefaults(config map[string]interface{}) map[string]interface{} {
	withConsolidation := ApplyAutoConsolidationDefault(config, AutoConsolidationEnvValue())
	withDispatch, _ := ApplyDevAgentHiveMindConfig(withConsolidation, DevAgentHiveMindEnvValue())
	return ApplyCoderAutonomousConfig(withDispatch)
}

// InteractionDefaults merges the interaction-time coder defaults into a
// dispatch config: chat-start tool-init seeding, the per-chat sandbox
// repository and branch, and the dev set-up. The core records the tool-call
// budget snapshot separately.
func InteractionDefaults(config map[string]interface{}, tools, tags []string, toolInit, overrides map[string]interface{}) map[string]interface{} {
	out := WithPreCallInit(config, tools, tags, DefaultPreCallInitResolvers())
	if repo := AutoSandboxRepoForInteraction(out, tags, toolInit, overrides); repo != "" {
		out = WithAutoSandboxRepo(out, repo)
	}
	out = WithDevSetup(out, DevSetupRepoForInteraction(out, tags, toolInit, overrides), DevSetupEnvValue())
	if branch := AutoSandboxBranchForInteraction(tags, toolInit); branch != "" {
		out = WithAutoSandboxBranch(out, branch)
	}
	return out
}

// ChatSetup applies the dev set-up default to a coder chat created directly
// (outside the dispatch path). A non-coder config is returned unchanged.
func ChatSetup(config map[string]interface{}, tags []string, toolInit map[string]interface{}) map[string]interface{} {
	if !IsCoderAgentConfig(config) {
		return config
	}
	return WithDevSetup(
		config,
		DevSetupRepoForInteraction(config, tags, toolInit, nil),
		DevSetupEnvValue(),
	)
}

// InteractionTags returns the tags a coder interaction should carry on top of
// the caller's explicit ones.
func InteractionTags(config map[string]interface{}, botName string, explicit []string) []string {
	return DefaultInteractionTags(config, botName, explicit)
}
