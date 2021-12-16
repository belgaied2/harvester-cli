package main

import (
	"os"
	"regexp"
	"strings"

	"github.com/belgaied2/harvester-cli/cmd"
	"github.com/pkg/errors"
	rcmd "github.com/rancher/cli/cmd"
	rancherprompt "github.com/rancher/cli/rancher_prompt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	if err := mainErr(); err != nil {
		logrus.Fatal(err)
	}
}

func mainErr() error {

	app := cli.NewApp()
	app.Name = "harvester"
	app.Usage = "Harvester CLI to easily manage infrastructure"
	// app.Before = func(ctx *cli.Context) error {
	// 	if ctx.GlobalBool("debug") {
	// 		logrus.SetLevel(logrus.DebugLevel)
	// 	}
	// 	return nil
	// }
	// app.Version = VERSION
	app.Author = "Mohamed Belgaied Hassine"
	app.Email = "mohamed.belgaiedhassine@gmail.com"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Debug logging",
		},
		cli.StringFlag{
			Name:   "config, conf",
			Usage:  "Path to Harvester's config file",
			EnvVar: "HARVESTER_CONFIG",
			Value:  os.ExpandEnv("${HOME}/.harvester/config"),
		},
	}
	app.Commands = []cli.Command{

		rcmd.LoginCommand(),
		cmd.ConfigCommand(),
		cmd.VMCommand(),
		cmd.ShellCommand(),
	}

	for _, com := range app.Commands {
		rancherprompt.Commands[com.Name] = com
		rancherprompt.Commands[com.ShortName] = com
	}
	rancherprompt.Flags = app.Flags
	parsed, err := parseArgs(os.Args)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	return app.Run(parsed)
}

var singleAlphaLetterRegxp = regexp.MustCompile("[a-zA-Z]")

func parseArgs(args []string) ([]string, error) {
	result := []string{}
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && len(arg) > 1 {
			for i, c := range arg[1:] {
				if string(c) == "=" {
					if i < 1 {
						return nil, errors.New("invalid input with '-' and '=' flag")
					}
					result[len(result)-1] = result[len(result)-1] + arg[i+1:]
					break
				} else if singleAlphaLetterRegxp.MatchString(string(c)) {
					result = append(result, "-"+string(c))
				} else {
					return nil, errors.Errorf("invalid input %v in flag", string(c))
				}
			}
		} else {
			result = append(result, arg)
		}
	}
	return result, nil
}
