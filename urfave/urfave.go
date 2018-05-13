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

package urfave

import (
	"regexp"
	"fmt"
	"github.com/joeycumines/cmd-doc/docgen"
	"strings"
)

type (
	// Help is urfave/cli help output parsed by heading.
	Help map[string]string

	commandHelp struct {
		base         []string
		name         string
		version      string
		usage        string
		buildDate    string
		buildUser    string
		buildVersion string
		description  string
		help         string
		commands     []docgen.Command
	}
)

func (h Help) Version() string {
	if h == nil {
		return ""
	}
	help, ok := h["Version"]
	if !ok {
		return ""
	}
	return ParseVersion(help)
}

func (h Help) Name() string {
	if h == nil {
		return ""
	}
	help, ok := h["Name"]
	if !ok {
		return ""
	}
	name, _ := ParseName(help)
	return name
}

func (h Help) Usage() string {
	if h == nil {
		return ""
	}
	help, ok := h["Name"]
	if !ok {
		return ""
	}
	_, usage := ParseName(help)
	return usage
}

func (h Help) Commands() []string {
	if h == nil {
		return nil
	}
	help, ok := h["Commands"]
	if !ok {
		return nil
	}
	return ParseCommands(help)
}

func (h Help) Description() string {
	if h == nil {
		return ""
	}
	help, ok := h["Description"]
	if !ok {
		return ""
	}
	return help
}

func (c commandHelp) Name() string {
	return c.name
}

func (c commandHelp) Info() string {
	var result string

	if v := strings.TrimSpace(c.name); v != "" {
		result += "name: " + v + "\n"
	}

	if v := strings.TrimSpace(c.version); v != "" {
		result += "version: " + v + "\n"
	}

	if v := strings.TrimSpace(c.buildVersion); v != "" {
		result += "build_version: " + v + "\n"
	}

	if v := strings.TrimSpace(c.buildDate); v != "" {
		result += "build_date: " + v + "\n"
	}

	if v := strings.TrimSpace(c.buildUser); v != "" {
		result += "build_user: " + v + "\n"
	}

	if v := strings.TrimSpace(c.usage); v != "" {
		if result == "" {
			result = v + "\n"
		} else {
			result = v + "\n\n" + result
		}
	}

	return result
}

func (c commandHelp) Description() string {
	return c.description
}

func (c commandHelp) Help() string {
	return c.help
}

func (c commandHelp) Commands() []docgen.Command {
	return c.commands
}

// generate recursively generates command help for each command and sub command
func (c *commandHelp) generate() error {
	helpStr, err := docgen.GetCommandOutput(append(c.base, "--help")...)

	if err != nil {
		return err
	}

	c.help = helpStr

	help := ParseHelp(helpStr)

	c.name = help.Name()

	c.version = help.Version()

	c.usage = help.Usage()

	c.buildVersion, _ = help["Build Version"]
	c.buildVersion = strings.TrimSpace(c.buildVersion)
	c.buildDate, _ = help["Build Date"]
	c.buildDate = strings.TrimSpace(c.buildDate)
	c.buildUser, _ = help["Build User"]
	c.buildUser = strings.TrimSpace(c.buildUser)

	c.description = ParseDescription(help.Description())

	for _, subCommand := range help.Commands() {
		command := commandHelp{
			base: make([]string, 0, len(c.base)+1),
		}
		command.base = append(command.base, c.base...)
		command.base = append(command.base, subCommand)

		if err := command.generate(); err != nil {
			return err
		}

		c.commands = append(c.commands, command)
	}

	return nil
}

// NewCommand generates a new command from urfave/cli help.
func NewCommand(command string, args ... string) (docgen.Command, error) {
	c := commandHelp{
		base: append([]string{command}, args...),
	}

	if err := c.generate(); err != nil {
		return nil, fmt.Errorf("urfave.NewCommand generate error: %s", err.Error())
	}

	return c, nil
}

func ParseCommands(commandsHelp string) []string {
	var result []string
	for _, line := range linesRegex.Split(commandsHelp, -1) {
		if sm := commandRegex.FindStringSubmatch(line); len(sm) >= 2 && sm[1] != "help" {
			result = append(result, sm[1])
		}
	}
	return result
}

func ParseName(nameHelp string) (name, usage string) {
	for _, line := range linesRegex.Split(nameHelp, -1) {
		split := strings.SplitN(line, " - ", 2)

		name = strings.TrimSpace(split[0])

		if name == "" {
			continue
		}

		if len(split) >= 2 {
			usage = strings.TrimSpace(split[1])
		}

		return
	}
	return
}

func ParseVersion(versionHelp string) string {
	for _, line := range linesRegex.Split(versionHelp, -1) {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func ParseDescription(descriptionHelp string) string {
	var result string

	for _, line := range linesRegex.Split(descriptionHelp, -1) {
		result += strings.TrimSpace(line) + "\n"
	}

	result = strings.TrimSpace(result)

	if result == "" {
		return ""
	}

	return result + "\n"
}

// ParseHelp splits the command help on all sections, mapping the name converted from format like `GLOBAL OPTIONS:`
// to format like `Global Options`, to the text segment.
func ParseHelp(help string) Help {
	var (
		header string
		result = make(map[string]string)
	)

	for _, line := range linesRegex.Split(help, -1) {
		if sm := headerRegex.FindStringSubmatch(line); len(sm) >= 2 {
			header = strings.Title(strings.ToLower(strings.TrimSpace(sm[1])))
			continue
		}

		if _, ok := result[header]; !ok {
			result[header] = ""
		}

		result[header] += line + "\n"
	}

	return result
}

var (
	linesRegex   = regexp.MustCompile(`[\n\r]`)
	headerRegex  = regexp.MustCompile(`^([A-Z]+[A-Z\s]*):\s*$`)
	commandRegex = regexp.MustCompile(`^\s+([A-Za-z][^\s]*?)[\s,]`)
)
