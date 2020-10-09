/*
   Copyright 2020 Docker Hub Tool authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package account

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/cli/cli/utils"
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"

	"github.com/docker/hub-cli-plugin/internal/format"
	"github.com/docker/hub-cli-plugin/internal/hub"
	"github.com/docker/hub-cli-plugin/internal/metrics"
)

const (
	infoName = "info"
)

type infoOptions struct {
	format.Option
}

func newInfoCmd(ctx context.Context, dockerCli command.Cli, parent string) *cobra.Command {
	var opts infoOptions
	cmd := &cobra.Command{
		Use:   infoName + " [OPTIONS]",
		Short: "Print the account information",
		Args:  cli.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, infoName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInfo(ctx, dockerCli, opts)
		},
	}
	opts.AddFormatFlag(cmd.Flags())
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func runInfo(ctx context.Context, dockerCli command.Cli, opts infoOptions) error {
	authResolver := func(hub *registry.IndexInfo) types.AuthConfig {
		return command.ResolveAuthConfig(ctx, dockerCli, hub)
	}
	client, err := hub.NewClient(authResolver)
	if err != nil {
		return err
	}
	user, err := client.GetUserInfo()
	if err != nil {
		return err
	}
	plan, err := client.GetHubPlan(user.ID)
	if err != nil {
		return err
	}
	return opts.Print(os.Stdout, account{user, plan}, printAccount)
}

func printAccount(out io.Writer, value interface{}) error {
	account := value.(account)
	// print user info
	w := tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, utils.Blue("Username:")+"\t%s\n", account.User.UserName)
	fmt.Fprintf(w, utils.Blue("Full name:")+"\t%s\n", account.User.FullName)
	fmt.Fprintf(w, utils.Blue("Company:")+"\t%s\n", account.User.Company)
	fmt.Fprintf(w, utils.Blue("Location:")+"\t%s\n", account.User.Location)
	fmt.Fprintf(w, utils.Blue("Joined:")+"\t%s ago\n", units.HumanDuration(time.Since(account.User.Joined)))
	fmt.Fprintf(w, utils.Blue("Plan:")+"\t%s\n", utils.Green(account.Plan.Name))
	if err := w.Flush(); err != nil {
		return err
	}

	// print plan info
	w = tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, utils.Blue("Limits:")+"\n")
	fmt.Fprintf(w, utils.Blue("%sSeats:")+"\t%v\n", pfx, account.Plan.Limits.Seats)
	fmt.Fprintf(w, utils.Blue("%sPrivate repositories:")+"\t%v\n", pfx, account.Plan.Limits.PrivateRepos)
	fmt.Fprintf(w, utils.Blue("%sParallel builds:")+"\t%v\n", pfx, account.Plan.Limits.ParallelBuilds)
	fmt.Fprintf(w, utils.Blue("%sCollaborators:")+"\t%v\n", pfx, getLimit(account.Plan.Limits.Collaborators))
	fmt.Fprintf(w, utils.Blue("%sTeams:")+"\t%v\n", pfx, getLimit(account.Plan.Limits.Teams))
	return w.Flush()
}

func getLimit(number int) string {
	if number == 9999 {
		return utils.Green("unlimited")
	}
	return fmt.Sprintf("%v", number)
}

const (
	pfx = "  "
)

type account struct {
	User *hub.User
	Plan *hub.Plan
}
