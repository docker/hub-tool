/*
   Copyright 2020 Docker Inc.

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

package format

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/pflag"
)

//Option handles format flags and printing the values depending the format
type Option struct {
	format string
}

//PrettyPrinter prints all the values in a pretty print format
type PrettyPrinter func(io.Writer, interface{}) error

//AddFormatFlag add the format flag to a command
func (o *Option) AddFormatFlag(flags *pflag.FlagSet) {
	flags.StringVar(&o.format, "format", "", `Print values using a custom format ("json")`)
}

//Print outputs values depending the given format
func (o *Option) Print(out io.Writer, values interface{}, prettyPrinter PrettyPrinter) error {
	switch o.format {
	case "":
		return prettyPrinter(out, values)
	case "json":
		return printJSON(out, values)
	default:
		return fmt.Errorf("unsupported format type: %q", o.format)
	}
}

func printJSON(out io.Writer, values interface{}) error {
	data, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, string(data))
	return err
}
