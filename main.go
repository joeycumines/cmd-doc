/*
   Copyright 2018 Joseph Cumines

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

package main

import (
	"gopkg.in/urfave/cli.v1"
	"os"
	"github.com/joeycumines/cmd-doc/urfave"
	"github.com/joeycumines/cmd-doc/docgen"
	"errors"
)

const (
	AppName        = `cmd-doc`
	AppUsage       = `generates markdown from commands`
	AppVersion     = `v0.0.0`
	AppDescription = `This command is a utility to allow automated documentation of binaries.

   Output will be printed to stdout in markdown format, and may be processed
   further, from there.`
)

type (
	Config struct {
		Header string
		Footer string
	}
)

func main() {
	app := cli.NewApp()

	app.Name = AppName
	app.Usage = AppUsage
	app.Version = AppVersion
	app.Description = AppDescription

	config := &Config{
	}

	app.Flags = append(
		app.Flags,
		cli.StringFlag{
			Name:        `header`,
			Usage:       `prepends to output as-is (no extra newline)`,
			Value:       config.Header,
			Destination: &config.Header,
		},
		cli.StringFlag{
			Name:        `footer`,
			Usage:       `appends to output as-is (no extra newline)`,
			Value:       config.Footer,
			Destination: &config.Footer,
		},
	)

	app.Commands = append(
		app.Commands,
		cli.Command{
			Name:  `urfave`,
			Usage: `outputs markdown from a golang command based on the urfave/cli package`,
			Description: `This command can be used to generate documentation from commands using
   the golang github.com/urfave/cli package, such as this one.

   Note that the command may be from anything runnable in the require format,
   including a dockerised binary.`,
			ArgsUsage: `[--] COMMAND [...ARGS]`,
			Action: func(c *cli.Context) error {
				var args []string = c.Args()

				if len(args) == 0 {
					return cli.NewExitError("command argument required", 1)
				}

				command, err := urfave.NewCommand(args[0], args[1:]...)

				if err != nil {
					return cli.NewExitError(err.Error(), 1)
				}

				err = config.Write(command)

				if err != nil {
					return cli.NewExitError(err.Error(), 1)
				}

				return nil
			},
		},
	)

	app.Run(os.Args)
}

func (c Config) Write(command docgen.Command) error {
	if command == nil {
		return errors.New("nil command")
	}
	if _, err := os.Stdout.WriteString(c.Header + docgen.GenerateMarkdown(command) + c.Footer); err != nil {
		return err
	}
	return nil
}
