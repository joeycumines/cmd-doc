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

package docgen

import (
	"strings"
	"os/exec"
	"fmt"
	"errors"
)

type (
	// Command models a command possibly with nested sub-commands.
	Command interface {
		// Name should be a single line like `some-command-name`.
		Name() string
		// Info is descriptive info about the command, each line will be block quoted.
		Info() string
		// Description is a text description, possibly with markdown formatting - it just gets copied right in.
		Description() string
		// Help the command help, it will be copied in as pre-formatted text.
		Help() string
		// Commands are any nested sub-commands.
		Commands() []Command
	}
)

// GetCommandOutput gets combined output as string from a slice, args len must be >= 1.
func GetCommandOutput(args ... string) (string, error) {
	if len(args) < 1 {
		return "", errors.New("docgen.GetCommandOutput requires at least one arg")
	}

	command := exec.Command(args[0], args[1:]...)

	b, err := command.CombinedOutput()

	if err != nil {
		err = fmt.Errorf(
			"docgen.GetCommandOutput exec error for args (%s): %s",
			strings.Join(args, ", "),
			err.Error(),
		)
	}

	return string(b), err
}

// GenerateMarkdown builds a markdown doc from a Command.
func GenerateMarkdown(command Command) string {
	var generate func(command Command, depth int) string
	generate = func(command Command, depth int) string {
		name, info, description, help, commands := command.Name(),
			command.Info(),
			command.Description(),
			command.Help(),
			command.Commands()

		// modify
		if name == "" {
			name = "COMMAND"
		}
		if lines := strings.Split(strings.TrimSuffix(info, "\n"), "\n"); len(lines) > 0 {
			info = ""
			for _, line := range lines {
				info += "> " + line + "\n"
			}
		}

		// newlines for optional segments
		if info != "" {
			info += "\n"
		}
		if description != "" {
			description += "\n"
		}

		// build
		body := strings.Repeat(`#`, depth+1) + " " + name + "\n" +
			"\n" +
			info +
			description +
			"```\n" +
			help + "\n" +
			"```\n"
		for _, subCommand := range commands {
			body += "\n" + generate(subCommand, depth+1)
		}

		return body
	}
	return generate(command, 0)
}
