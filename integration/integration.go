// Package communitycoder registers the public Community Coder integration: a
// thinned-down, publicly available version of the private msgmate_coder
// integration. It shares the coder turn hooks and policy with the private
// integration through the public `coder` library, and drives its turns through
// the core's generic tool loop (Host v5 Agent runtime).
//
// The community coder deliberately ships without the private integrations'
// optimisations (MCP endpoints, TUI/PTY relay, provider admission tuning,
// dev-setup fleet, dispatch/consolidation/lead) and without the git
// integration's custom bot-selector bots.
package communitycoder

import (
	"context"

	"github.com/msgmate-io/go-integration-interface/integrationinterface"
	"github.com/msgmate-io/open-chat-go-community-coder-integration/coder"
)

// ChatBackendName is the shared-config `chat_backend` value that marks a chat
// as a community coder chat.
const ChatBackendName = "community_coder"

const readmeMarkdown = `# Community Coder

A thinned-down, publicly available coding agent for Open-Chat. It shares the
coder turn hooks and policy with the private ` + "`msgmate_coder`" + ` integration
through the public ` + "`coder`" + ` library and runs its turns through the core's
generic tool loop.
`

func init() {
	integrationinterface.MustRegister(integrationinterface.Definition{
		Name:           "community_coder",
		AdminOnly:      false,
		UserAccessible: true,
		ReadmeMarkdown: readmeMarkdown,
		// The community coder shares the agent config defaults with the private
		// coder integration through the public coder library.
		AgentConfig:              coder.AgentConfig,
		AgentDefaults:            coder.BotDefaults,
		AgentInteractionDefaults: coder.InteractionDefaults,
		AgentChatSetup:           coder.ChatSetup,
		AgentInteractionTags:     coder.InteractionTags,
		CLICommands:              codeCLICommands(),
	})
	integrationinterface.RegisterChatBackend(integrationinterface.ChatBackendSpec{
		Name:             ChatBackendName,
		Backend:          communityAgentBackend,
		StateProvider:    communityStateProvider,
		InterruptHandler: communityInterruptHandler,
	})
}

// communityAgentBackend runs a community coder turn through the core Agent
// runtime with the shared coder hooks.
func communityAgentBackend(ctx context.Context, req integrationinterface.ChatBackendRequest) error {
	_, err := integrationinterface.RequireHost().Agent().Run(ctx, integrationinterface.AgentRequest{
		ChatUUID: req.Message.ChatUUID,
		APIHost:  req.APIHost,
		BotUser:  req.BotUser,
		Message:  req.Message,
		Messages: req.Messages,
		Config:   req.Config,
		// The integration owns the coder system prompt via the shared coder lib.
		SystemPrompt: coder.ComposeSystemPrompt(req.Config),
		Hooks:        coder.TurnHooks(),
	})
	return err
}

func communityStateProvider(chatUUID string) (integrationinterface.BackendState, bool) {
	if integrationinterface.RequireHost().Chat().HasActiveGeneration(chatUUID) {
		return integrationinterface.BackendStateRunning, true
	}
	return integrationinterface.BackendStateIdle, true
}

func communityInterruptHandler(chatUUID string) bool {
	return integrationinterface.RequireHost().Chat().CancelActiveGeneration(chatUUID)
}
