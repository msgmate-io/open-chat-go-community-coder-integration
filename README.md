# Open-Chat Community Coder Integration

A thinned-down, publicly available coding agent integration for
[Open-Chat](https://github.com/msgmate-io/open-chat-go).

It is the public sibling of the private `msgmate_coder` integration and shares
the coder turn hooks and policy with it through the public `coder` library in
this repository, so the two cannot drift apart.

## Layout

- `coder/` — the shared, dependency-free coder policy library (Host v5 Agent
  turn hooks: workspace-root seeding, `coder_bash` tolerance rewrite,
  sandbox-failure framing, bookkeeping-tool budget exemption). Any Open-Chat
  coder integration, public or private, may import it.
- `integration/` — registers the `community_coder` integration and its chat
  backend, which runs turns through the core's generic tool loop (Host v5 Agent
  runtime).

## CLI

When the TUI integration is also part of the build, the community coder
contributes the `open-chat code` alias. It launches the TUI and forwards a
`--bot <username>` selection; the private `msgmate_coder` integration provides
a richer `code` command (local-folder sandbox + bound chat) and takes precedence
when both are compiled in.

## Profile

The community coder is built through the public `community-coder` integration
profile. It is never built together with the private `msgmate_coder` integration
(they share the `coder_*` Go tool names); the private profiles exclude it with
`!community_coder`.
