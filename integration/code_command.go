package communitycoder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/msgmate-io/go-integration-interface/integrationinterface"
	"github.com/urfave/cli/v3"
)

// codeCLICommand contributes the public `open-chat code` alias. The community
// coder has no server-side workspace/session API (it runs through the core's
// generic tool loop), so the alias simply opens the TUI, optionally pinned to a
// specific bot. The private msgmate_coder integration contributes its own,
// richer `code` command and takes precedence when both are compiled in.
func codeCLICommand() *cli.Command {
	return &cli.Command{
		Name:  "code",
		Usage: "Open a coding chat in the Open-Chat TUI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "bot",
				Usage: "Bot username to open the coding chat with",
				Value: "",
			},
			&cli.StringSliceFlag{
				Name:      "arg",
				Usage:     "Extra argument forwarded to `open-chat tui`",
				TakesFile: false,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			executable, err := os.Executable()
			if err != nil {
				return err
			}
			args := []string{"tui"}
			if bot := strings.TrimSpace(c.String("bot")); bot != "" {
				args = append(args, "--bot", bot)
			}
			args = append(args, c.StringSlice("arg")...)
			cmd := exec.CommandContext(ctx, executable, args...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("launching the Open-Chat TUI: %w", err)
			}
			return nil
		},
	}
}

// codeCLICommands declares the alias lazily: the command builder runs after all
// integration init functions, so it can require the TUI integration to be part
// of the build instead of only failing at launch time.
func codeCLICommands() []integrationinterface.CLICommand {
	return []integrationinterface.CLICommand{{
		Name: "code",
		Command: func() *cli.Command {
			if !integrationinterface.Has("tui") {
				return nil
			}
			return codeCLICommand()
		},
	}}
}
