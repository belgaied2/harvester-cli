package cmd

import (
	"context"

	rcmd "github.com/rancher/cli/cmd"
	"github.com/urfave/cli"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TemplateData struct {
	Name          string
	LatestVersion int
}

// TemplateCommand defines the CLI command that lists VM templates in Harvester
func TemplateCommand() cli.Command {
	return cli.Command{
		Name:    "template",
		Aliases: []string{"tpl"},
		Usage:   "Manipulate VM templates",
		Action:  templateList,
		Flags: []cli.Flag{
			nsFlag,
		},
		Subcommands: cli.Commands{
			cli.Command{
				Name:        "list",
				Aliases:     []string{"ls"},
				Usage:       "List templates",
				Description: "\nLists all the VM templates available in Harvester",
				ArgsUsage:   "None",
				Action:      templateList,
				Flags: []cli.Flag{
					nsFlag,
				},
			},
		},
	}
}

func templateList(ctx *cli.Context) (err error) {
	c, err := GetHarvesterClient(ctx)

	if err != nil {
		return
	}

	tplList, err := c.HarvesterhciV1beta1().VirtualMachineTemplates(ctx.String("namespace")).List(context.TODO(), k8smetav1.ListOptions{})

	if err != nil {
		return
	}

	writer := rcmd.NewTableWriter([][]string{
		{"NAME", "Name"},
		{"LATEST_VERSION", "LatestVersion"},
	},
		ctx)

	defer writer.Close()

	for _, tplItem := range tplList.Items {

		writer.Write(&TemplateData{
			Name:          tplItem.Name,
			LatestVersion: tplItem.Status.LatestVersion,
		})

	}

	return writer.Err()
}
