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
	"golang.org/x/sync/errgroup"

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
		Use:                   infoName + " [OPTIONS] [ORGANIZATION]",
		Short:                 "Print the account information",
		Args:                  cli.RequiresMaxArgs(1),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			"sudo": "true",
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, infoName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return runOrgInfo(streams, hubClient, opts, args[0])
			}
			return runUserInfo(streams, hubClient, opts)
		},
	}
	opts.AddFormatFlag(cmd.Flags())
	return cmd
}

func runOrgInfo(streams command.Streams, hubClient *hub.Client, opts infoOptions, orgName string) error {
	var (
		org         *hub.Account
		consumption *hub.Consumption
	)

	g := errgroup.Group{}
	g.Go(func() error {
		var err error
		org, err = hubClient.GetOrganizationInfo(orgName)
		return checkForbiddenError(err)
	})
	g.Go(func() error {
		var err error
		consumption, err = hubClient.GetOrgConsumption(orgName)
		return checkForbiddenError(err)
	})
	if err := g.Wait(); err != nil {
		return err
	}

	plan, err := hubClient.GetHubPlan(org.ID)
	if err != nil {
		return checkForbiddenError(err)
	}

	return opts.Print(streams.Out(), account{org, plan, consumption}, printAccount)
}

func runUserInfo(streams command.Streams, hubClient *hub.Client, opts infoOptions) error {
	user, err := hubClient.GetUserInfo()
	if err != nil {
		return checkForbiddenError(err)
	}
	consumption, err := hubClient.GetUserConsumption(user.Name)
	if err != nil {
		return checkForbiddenError(err)
	}
	plan, err := hubClient.GetHubPlan(user.ID)
	if err != nil {
		return checkForbiddenError(err)
	}

	return opts.Print(streams.Out(), account{user, plan, consumption}, printAccount)
}

func checkForbiddenError(err error) error {
	if hub.IsForbiddenError(err) {
		return fmt.Errorf(ansi.Error("failed to get organization information, you need to be the organization Owner"))
	}
	return err
}

func printAccount(out io.Writer, value interface{}) error {
	account := value.(account)

	// print user info
	fmt.Fprintf(out, ansi.Key("Name:")+"\t\t%s\n", account.Account.Name)
	fmt.Fprintf(out, ansi.Key("Full name:")+"\t%s\n", account.Account.FullName)
	fmt.Fprintf(out, ansi.Key("Company:")+"\t%s\n", account.Account.Company)
	fmt.Fprintf(out, ansi.Key("Location:")+"\t%s\n", account.Account.Location)
	fmt.Fprintf(out, ansi.Key("Joined:")+"\t\t%s ago\n", units.HumanDuration(time.Since(account.Account.Joined)))
	fmt.Fprintf(out, ansi.Key("Plan:")+"\t\t%s\n", ansi.Emphasise(account.Plan.Name))

	// print plan info
	fmt.Fprintf(out, ansi.Key("Limits:")+"\n")
	fmt.Fprintf(out, ansi.Key("  Seats:")+"\t\t%v\n", getCurrentLimit(account.Consumption.Seats, account.Plan.Limits.Seats))
	fmt.Fprintf(out, ansi.Key("  Private repositories:")+"\t%v\n", getCurrentLimit(account.Consumption.PrivateRepositories, account.Plan.Limits.PrivateRepos))
	fmt.Fprintf(out, ansi.Key("  Teams:")+"\t\t%v\n", getCurrentLimit(account.Consumption.Teams, account.Plan.Limits.Teams))
	fmt.Fprintf(out, ansi.Key("  Collaborators:")+"\t%v\n", getLimit(account.Plan.Limits.Collaborators))
	fmt.Fprintf(out, ansi.Key("  Parallel builds:")+"\t%v\n", getLimit(account.Plan.Limits.ParallelBuilds))

	return nil
}

func getCurrentLimit(current, limit int) string {
	if limit == 9999 {
		return ansi.Emphasise("unlimited")
	}
	return fmt.Sprintf("%v/%v", current, limit)
}

func getLimit(limit int) string {
	if limit == 9999 {
		return ansi.Emphasise("unlimited")
	}
	return fmt.Sprintf("%v", limit)
}

type account struct {
	Account     *hub.Account
	Plan        *hub.Plan
	Consumption *hub.Consumption
}
