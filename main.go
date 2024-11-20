package main

import (
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/belgaied2/harvester-cli/cmd"
	"github.com/belgaied2/harvester-cli/constant"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type AppContext struct {
	UserHomeDirectory string
	AppObject         *cli.App
}

func main() {
	app := AppContext{
		UserHomeDirectory: getUserHomeDirectory(),
		AppObject:         cli.NewApp(),
	}

	setAppMetadata(app)
	setAppFlags(app)
	setAppCommands(app)

	arguments, err := parseArgs(os.Args)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	err = app.AppObject.Run(arguments)
	if err != nil {
		logrus.Fatal(err)
	}
}

func getUserHomeDirectory() string {
	userHome, err := os.UserHomeDir()
	if err != nil {
		logrus.Warn("Not able to determine home folder of current user!")
	}
	return userHome
}

func setAppMetadata(app AppContext) {
	app.AppObject.Name = constant.Name
	app.AppObject.Usage = constant.Description
	app.AppObject.Version = constant.Version
	app.AppObject.Authors = constant.Authors
	app.AppObject.EnableBashCompletion = true
}

func setAppCommands(app AppContext) {
	app.AppObject.Commands = []*cli.Command{
		cmd.LoginCommand(),
		cmd.ConfigCommand(),
		cmd.VMCommand(),
		cmd.ShellCommand(),
		cmd.TemplateCommand(),
		cmd.ImageCommand(),
		cmd.KeypairCommand(),
		cmd.ImportCommand(),
		cmd.CompleteCommand(),
	}
}

func setAppFlags(app AppContext) {
	app.AppObject.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Debug logging",
		},
		&cli.StringFlag{
			Name:    "harvester-config, hconf",
			Usage:   "Path to Harvester's config file",
			EnvVars: []string{"HARVESTER_CONFIG"},
			Value:   path.Join(app.UserHomeDirectory, ".harvester", "config"),
		},
		&cli.StringFlag{
			Name:    "config, rconf",
			Usage:   "Path to Rancher's config file",
			EnvVars: []string{"RANCHER_CONFIG"},
			Value:   path.Join(app.UserHomeDirectory, ".rancher"),
		},
	}
}

func parseArgs(args []string) ([]string, error) {
	var result []string

	for _, arg := range args {
		if isShortFlag(arg) {
			parsed, err := parseShortFlag(arg)
			if err != nil {
				return nil, err
			}
			result = append(result, parsed...)
		} else {
			result = append(result, arg)
		}
	}

	return result, nil
}

func isShortFlag(arg string) bool {
	return strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && len(arg) > 1
}

func parseShortFlag(flag string) ([]string, error) {
	var result []string

	for index, raw_char := range flag[1:] {
		char := string(raw_char)

		if char == "=" {
			if index == 0 {
				return nil, errors.New("invalid input: '-' cannot be directly followed by '='")
			}
			result[len(result)-1] += flag[index+2:]
			break
		} else if regexp.MustCompile("^[a-zA-Z]$").MatchString(char) {
			result = append(result, "-"+char)
		} else {
			return nil, errors.New("invalid input: unexpected character in flag: " + char)
		}
	}

	return result, nil
}
