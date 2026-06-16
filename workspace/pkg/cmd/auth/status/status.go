package status

import (
	"fmt"
	"net/http"

	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

type StatusOptions struct {
	IO         *iostreams.IOStreams
	Config     func() (config.Config, error)
	HttpClient func() (*http.Client, error)
	Hostname   string
	ShowToken  bool
}

func NewCmdStatus(f *cmdutil.Factory, runF func(*StatusOptions) error) *cobra.Command {
	opts := &StatusOptions{
		IO:         f.IOStreams,
		Config:     f.Config,
		HttpClient: f.HttpClient,
	}

	cmd := &cobra.Command{
		Use:   "status",
		Args:  cobra.NoArgs,
		Short: "View authentication status",
		Long:  `Verifies and displays information about your authentication state.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return runStatus(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Hostname, "hostname", "h", "", "Check a specific hostname's auth status")
	cmd.Flags().BoolVarP(&opts.ShowToken, "show-token", "t", false, "Display the auth token")

	return cmd
}

func runStatus(opts *StatusOptions) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	var hosts []string
	if opts.Hostname != "" {
		hosts = []string{opts.Hostname}
	} else {
		hosts = cfg.Hosts()
	}

	if len(hosts) == 0 {
		cs := opts.IO.ColorScheme()
		fmt.Fprintf(opts.IO.ErrOut, "You are not logged into any GitHub hosts. Run %s to authenticate.\n", cs.Bold("gh auth login"))
		return cmdutil.SilentError
	}

	stdout := opts.IO.Out
	stderr := opts.IO.ErrOut
	cs := opts.IO.ColorScheme()

	var failed bool
	for _, host := range hosts {
		token, source := cfg.Get(host, "oauth_token")
		if token == "" {
			fmt.Fprintf(stderr, "%s: not logged in\n", host)
			failed = true
			continue
		}

		client, err := opts.HttpClient()
		if err != nil {
			return err
		}

		username, err := api.CurrentUsername(client, host)
		if err != nil {
			fmt.Fprintf(stderr, "%s %s: Active directory/API error: %s\n", cs.Red("X"), host, err)
			failed = true
			continue
		}

		fmt.Fprintf(stdout, "%s %s: logged in as %s (%s)\n", cs.Green("✓"), host, cs.Bold(username), source)
		if opts.ShowToken {
			fmt.Fprintf(stdout, "  - Token: %s\n", token)
		}
	}

	if failed {
		return cmdutil.SilentError
	}

	return nil
}
