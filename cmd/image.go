package cmd

import (
	"context"

	rcmd "github.com/rancher/cli/cmd"
	"github.com/urfave/cli"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ImageData struct {
	Name       string
	Id         string
	SourceType string
	Url        string
}

// TemplateCommand defines the CLI command that lists VM templates in Harvester
func ImageCommand() cli.Command {
	return cli.Command{
		Name:    "image",
		Aliases: []string{"img"},
		Usage:   "Manipulate VM images",
		Action:  imageList,
		Flags: []cli.Flag{
			nsFlag,
		},
		Subcommands: cli.Commands{
			cli.Command{
				Name:        "list",
				Aliases:     []string{"ls"},
				Usage:       "List VM images",
				Description: "\nLists all the VM images available in Harvester",
				ArgsUsage:   "None",
				Action:      imageList,
				Flags: []cli.Flag{
					nsFlag,
				},
			},
		},
	}
}

func imageList(ctx *cli.Context) (err error) {
	c, err := GetHarvesterClient(ctx)

	if err != nil {
		return
	}

	imgList, err := c.HarvesterhciV1beta1().VirtualMachineImages(ctx.String("namespace")).List(context.TODO(), k8smetav1.ListOptions{})

	if err != nil {
		return
	}

	writer := rcmd.NewTableWriter([][]string{
		{"NAME", "Name"},
		{"ID", "Id"},
		{"SOURCE TYPE", "SourceType"},
		{"URL", "Url"},
	},
		ctx)

	defer writer.Close()

	for _, imgItem := range imgList.Items {

		writer.Write(&ImageData{
			Name:       imgItem.Spec.DisplayName,
			Id:         imgItem.Name,
			SourceType: imgItem.Spec.SourceType,
			Url:        imgItem.Spec.URL,
		})

	}

	return writer.Err()
}
