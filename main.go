package main

import (
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/belgaied2/harvester-cli/cmd"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var VERSION = "dev"

func main() {
	if err := mainErr(); err != nil {
		logrus.Fatal(err)
	}
}

func mainErr() error {

	userHome, err := os.UserHomeDir()

	if err != nil {
		logrus.Warn("Not able to determine home folder of current user!")
	}

	app := cli.NewApp()
	app.Name = "harvester"
	app.Usage = "Harvester CLI to easily manage infrastructure"
	// app.Before = func(ctx *cli.Context) error {
	// 	if ctx.GlobalBool("debug") {
	// 		logrus.SetLevel(logrus.DebugLevel)
	// 	}
	// 	return nil
	// }
	app.Version = VERSION
	app.Authors = append(app.Authors, &cli.Author{
		Name:  "Mohamed Belgaied Hassine",
		Email: "mohamed.belgaiedhassine@gmail.com"})
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Debug logging",
		},
		&cli.StringFlag{
			Name:    "harvester-config, hconf",
			Usage:   "Path to Harvester's config file",
			EnvVars: []string{"HARVESTER_CONFIG"},
			Value:   path.Join(userHome, ".harvester", "config"),
		},
		&cli.StringFlag{
			Name:    "config, rconf",
			Usage:   "Path to Rancher's config file",
			EnvVars: []string{"RANCHER_CONFIG"},
			Value:   path.Join(userHome, ".rancher"),
		},
		// cli.StringFlag{
		// 	Name:   "loglevel",
		// 	Usage:  "Defines the log level to be used, possible values are error, info, warn, debug and trace",
		// 	EnvVar: "HARVESTER_LOG",
		// 	Value:  "info",
		// },
	}
	app.Commands = []*cli.Command{

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
	app.EnableBashCompletion = true

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
