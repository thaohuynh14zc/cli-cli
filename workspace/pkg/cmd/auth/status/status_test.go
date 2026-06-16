package status

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func Test_runStatus(t *testing.T) {
	tests := []struct {
		name       string
		opts       *StatusOptions
		isTTY      bool
		wantOut    string
		wantErrOut string
		wantErr    bool
	}{
		{
			name: "no hosts",
			opts: &StatusOptions{
				Config: func() (config.Config, error) {
					return config.NewBlankConfig(), nil
				},
			},
			wantErr:    true,
			wantErrOut: "You are not logged into any GitHub hosts. Run gh auth login to authenticate.\n",
		},
		{
			name: "show token",
			opts: &StatusOptions{
				ShowToken: true,
				Config: func() (config.Config, error) {
					cfg := config.NewBlankConfig()
					_ = cfg.Set("github.com", "oauth_token", "oauth_token")
					return cfg, nil
				},
				HttpClient: func() (*http.Client, error) {
					reg := &httpmock.Registry{}
					reg.Register(
						httpmock.GraphQL(`query UserCurrent\b`),
						httpmock.StringResponse(`{"data":{"viewer":{"login":"hubot"}}}`),
					)
					return &http.Client{Transport: reg}, nil
				},
			},
			wantOut: "✓ github.com: logged in as hubot (config)\n  - Token: oauth_token\n",
		},
		{
			name: "show token invalid",
			opts: &StatusOptions{
				ShowToken: true,
				Config: func() (config.Config, error) {
					cfg := config.NewBlankConfig()
					_ = cfg.Set("github.com", "oauth_token", "oauth_token")
					return cfg, nil
				},
				HttpClient: func() (*http.Client, error) {
					reg := &httpmock.Registry{}
					reg.Register(
						httpmock.GraphQL(`query UserCurrent\b`),
						httpmock.StatusStringResponse(401, `{"message":"Bad credentials"}`),
					)
					return &http.Client{Transport: reg}, nil
				},
			},
			wantErr:    true,
			wantErrOut: "X github.com: Active directory/API error: HTTP 401: Bad credentials (https://api.github.com/graphql)\n",
		},
		{
			name: "hostname not logged in",
			opts: &StatusOptions{
				Hostname: "github.com",
				Config: func() (config.Config, error) {
					return config.NewBlankConfig(), nil
				},
			},
			wantErr:    true,
			wantErrOut: "github.com: not logged in\n",
		},
		{
			name: "logged in",
			opts: &StatusOptions{
				Config: func() (config.Config, error) {
					cfg := config.NewBlankConfig()
					_ = cfg.Set("github.com", "oauth_token", "oauth_token")
					return cfg, nil
				},
				HttpClient: func() (*http.Client, error) {
					reg := &httpmock.Registry{}
					reg.Register(
						httpmock.GraphQL(`query UserCurrent\b`),
						httpmock.StringResponse(`{"data":{"viewer":{"login":"hubot"}}}`),
					)
					return &http.Client{Transport: reg}, nil
				},
			},
			wantOut: "✓ github.com: logged in as hubot (config)\n",
		},
		{
			name: "token invalid",
			opts: &StatusOptions{
				Config: func() (config.Config, error) {
					cfg := config.NewBlankConfig()
					_ = cfg.Set("github.com", "oauth_token", "oauth_token")
					return cfg, nil
				},
				HttpClient: func() (*http.Client, error) {
					reg := &httpmock.Registry{}
					reg.Register(
						httpmock.GraphQL(`query UserCurrent\b`),
						httpmock.StatusStringResponse(401, `{"message":"Bad credentials"}`),
					)
					return &http.Client{Transport: reg}, nil
				},
			},
			wantErr:    true,
			wantErrOut: "X github.com: Active directory/API error: HTTP 401: Bad credentials (https://api.github.com/graphql)\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			io, _, stdout, stderr := iostreams.Test()
			io.SetStdoutTTY(tt.isTTY)
			io.SetStderrTTY(tt.isTTY)

			opts := tt.opts
			opts.IO = io

			err := runStatus(opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("runStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.wantOut, stdout.String())
			assert.Equal(t, tt.wantErrOut, stderr.String())
		})
	}
}
