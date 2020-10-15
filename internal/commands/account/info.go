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
	"github.com/juju/ansiterm"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/color"
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
	w := ansiterm.NewTabWriter(out, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, color.Key("Username:")+"\t%s\n", account.User.UserName)
	fmt.Fprintf(w, color.Key("Full name:")+"\t%s\n", account.User.FullName)
	fmt.Fprintf(w, color.Key("Company:")+"\t%s\n", account.User.Company)
	fmt.Fprintf(w, color.Key("Location:")+"\t%s\n", account.User.Location)
	fmt.Fprintf(w, color.Key("Joined:")+"\t%s ago\n", units.HumanDuration(time.Since(account.User.Joined)))
	fmt.Fprintf(w, color.Key("Plan:")+"\t%s\n", color.Emphasise(account.Plan.Name))
	if err := w.Flush(); err != nil {
		return err
	}

	// print plan info
	w = ansiterm.NewTabWriter(out, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, color.Key("Limits:")+"\n")
	fmt.Fprintf(w, color.Key("%sSeats:")+"\t%v\n", pfx, account.Plan.Limits.Seats)
	fmt.Fprintf(w, color.Key("%sPrivate repositories:")+"\t%v\n", pfx, account.Plan.Limits.PrivateRepos)
	fmt.Fprintf(w, color.Key("%sParallel builds:")+"\t%v\n", pfx, account.Plan.Limits.ParallelBuilds)
	fmt.Fprintf(w, color.Key("%sCollaborators:")+"\t%v\n", pfx, getLimit(account.Plan.Limits.Collaborators))
	fmt.Fprintf(w, color.Key("%sTeams:")+"\t%v\n", pfx, getLimit(account.Plan.Limits.Teams))
	return w.Flush()
}

func getLimit(number int) string {
	if number == 9999 {
		return color.Emphasise("unlimited")
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
