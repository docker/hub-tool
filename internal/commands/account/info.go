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
	"fmt"
	"io"
	"time"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/internal/format"
	"github.com/docker/hub-tool/internal/hub"
	"github.com/docker/hub-tool/internal/metrics"
)

const (
	infoName = "info"
)

type infoOptions struct {
	format.Option
}

func newInfoCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts infoOptions
	cmd := &cobra.Command{
		Use:   infoName + " [OPTIONS]",
		Short: "Print the account information",
		Args:  cli.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, infoName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInfo(streams, hubClient, opts)
		},
	}
	opts.AddFormatFlag(cmd.Flags())
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func runInfo(streams command.Streams, hubClient *hub.Client, opts infoOptions) error {
	user, err := hubClient.GetUserInfo()
	if err != nil {
		return err
	}
	plan, err := hubClient.GetHubPlan(user.ID)
	if err != nil {
		return err
	}
	return opts.Print(streams.Out(), account{user, plan}, printAccount)
}

func printAccount(out io.Writer, value interface{}) error {
	account := value.(account)

	// print user info
	fmt.Printf(ansi.Key("Username:")+"\t%s\n", account.User.UserName)
	fmt.Printf(ansi.Key("Full name:")+"\t%s\n", account.User.FullName)
	fmt.Printf(ansi.Key("Company:")+"\t%s\n", account.User.Company)
	fmt.Printf(ansi.Key("Location:")+"\t%s\n", account.User.Location)
	fmt.Printf(ansi.Key("Joined:")+"\t\t%s ago\n", units.HumanDuration(time.Since(account.User.Joined)))
	fmt.Printf(ansi.Key("Plan:")+"\t\t%s\n", ansi.Emphasise(account.Plan.Name))

	// print plan info
	fmt.Printf(ansi.Key("Limits:") + "\n")
	fmt.Printf(ansi.Key("  Seats:")+"\t\t%v\n", account.Plan.Limits.Seats)
	fmt.Printf(ansi.Key("  Private repositories:")+"\t%v\n", account.Plan.Limits.PrivateRepos)
	fmt.Printf(ansi.Key("  Parallel builds:")+"\t%v\n", account.Plan.Limits.ParallelBuilds)
	fmt.Printf(ansi.Key("  Collaborators:")+"\t%v\n", getLimit(account.Plan.Limits.Collaborators))
	fmt.Printf(ansi.Key("  Teams:")+"\t\t%v\n", getLimit(account.Plan.Limits.Teams))

	return nil
}

func getLimit(number int) string {
	if number == 9999 {
		return ansi.Emphasise("unlimited")
	}
	return fmt.Sprintf("%v", number)
}

type account struct {
	User *hub.User
	Plan *hub.Plan
}
